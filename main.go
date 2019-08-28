/*
 * This program compares the content of two directories
 *
 * TODO export some functionality into classes
 */


package main

import (
  "fmt"
  "sync"
  "os"
  "time"
)


func process(sem chan bool, dir *Dir, wg *sync.WaitGroup) {
  // Decreasing internal counter for wait-group as soon as goroutine finishes
  defer wg.Done()

  // Sending increments, reading decrements semaphore
  sem <- true
  defer func() { <-sem }()

  // Process directory
  processDir(dir)

  // Process children
  for _, d := range dir.dirs.both {
    wg.Add(1)
    go process( sem, d, wg )
  }
  // Logging
  fmt.Print(dir.formatChanges())
}

func main() {
  wg := new(sync.WaitGroup)

  // initialize folders to compare
  if len(os.Args) == 3 {
    SetPrefixes(os.Args[1],os.Args[2])
    fmt.Printf("Comparing \"%s\" to \"%s\"\n", os.Args[1], os.Args[2])
  } else {
    fmt.Println("Please specify the folders to compare as arguments.")
    return
  }
  dir := NewDirRoot()

  // Create semaphore to limit number of workers
  // Without this limit, OS runs out of file descriptors and thus returns invalid results
  sem := make(chan bool, 8)

  // Start measuring time
  start := time.Now()

  // Process directory
  wg.Add(1)
  go process(sem, dir, wg)

  // Waiting for all goroutines to finish (otherwise they die as main routine dies)
  wg.Wait()

  // Calculate elapsed time
  elapsed := time.Since(start)
  fmt.Printf("Elapsed time %s\n", elapsed)

  // Propagate directory state up to root
  fmt.Println("Propagate states")
  dir.updateStates()

  // Inspect result
  // NOTE this fills up several GB of RAM for big folders within very short time!
  // -> better use print in process() or save to file
  //fmt.Println("Inspect result")
  //dir.printRecursive()
}

