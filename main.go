package main

import (
  "fmt"
  "io/ioutil"
  "path/filepath"
  "sync"
  "os"
  "log"
  "time"
)

var prefixes [2]string

// NOTE ioutil.ReadDir() returns entries SORTED by filename (https://golang.org/pkg/io/ioutil/#ReadDir)
type FilePair struct{ name string; f0 os.FileInfo; f1 os.FileInfo }
func compareNodes(A, B []os.FileInfo) (onlyA, onlyB []os.FileInfo, both []FilePair) {
  var i, j int
  for i < len(A) && j < len(B) {
    if A[i].Name() < B[j].Name() {
      fmt.Println("[A]"+A[i].Name())
      onlyA = append(onlyA, A[i])
      i++
    } else if A[i].Name() > B[j].Name() {
      fmt.Println("[B]"+B[j].Name())
      onlyB = append(onlyB, B[j])
      j++
    } else { //A[i] == B[j]
      //fmt.Println("[~]"+A[i].Name())
      both = append(both, FilePair{A[i].Name(),A[i],B[j]})
      i++
      j++
    }
  }
  return
}

func compareFiles(dir string, f0, f1 os.FileInfo) bool {
  if f0.Size() != f1.Size() { return false }
  //if f0.ModTime() != f1.ModTime() { return false }
  // TODO compare file content: https://stackoverflow.com/a/30038571/3779655
  return true
}

func process(sem chan bool, dir string, wg *sync.WaitGroup) {
  // Decreasing internal counter for wait-group as soon as goroutine finishes
  defer wg.Done()

  // Sending increments, reading decrements semaphore
  sem <- true
  defer func() { <-sem }()

  // Process directory
  // read nodes
  var nodes [2][]os.FileInfo
  for i := 0; i < 2; i++ {
    nodes[i], _ = ioutil.ReadDir(filepath.Join(prefixes[i],dir))
  }
  // compare node names
  _,_,b := compareNodes(nodes[0], nodes[1])
  // inspect result
  for _, n := range b {
    if n.f0.IsDir() && n.f1.IsDir() {
      // recursively process dirs present on both sides
      wg.Add(1)
      go process(sem, filepath.Join(dir, n.name), wg)
    }
    if !n.f0.IsDir() && !n.f1.IsDir() {
      // compare files
      if !compareFiles(dir, n.f0, n.f1) {
        fmt.Println("[!]"+n.name)
      }
    }
  }
}

func main() {
  // Create wait group to wait until all workers are finished
  wg := new(sync.WaitGroup)

  // initialize folders to compare
  if len(os.Args) == 3 {
    prefixes[0] = os.Args[1]
    prefixes[1] = os.Args[2]
  } else {
    log.Fatal("Error: No folders to compare passed as arguments.")
  }

  // Create semaphore to limit number of workers
  sem := make(chan bool, 64)

  // Start measuring time
  start := time.Now()

  // Start traversing file tree
  wg.Add(1)
  go process(sem, ".", wg)

  // Waiting for all goroutines to finish (otherwise they die as main routine dies)
  wg.Wait()

  // Calculate elapsed time
  elapsed := time.Since(start)
  fmt.Printf("Elapsed time %s\n", elapsed)
}
