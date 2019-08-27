/*
 * This program compares the content of two directories
 *
 * TODO export some functionality into classes
 */


package main

import (
  "fmt"
  "io/ioutil"
  "path/filepath"
  "sync"
  "os"
)

var prefixes [2]string

func children2map(children []os.FileInfo) (dirs, files map[string]os.FileInfo) {
  dirs = make(map[string]os.FileInfo)
  files = make(map[string]os.FileInfo)
  for _, f := range children {
    if f.IsDir() {
      dirs[f.Name()] = f
    } else {
      files[f.Name()] = f
    }
  }
  return
}

type Filter func(f os.FileInfo)bool
func filter(a []os.FileInfo, keep Filter)[]os.FileInfo {
  n := 0
  for _, x := range a {
  	if keep(x) {
  		a[n] = x
  		n++
  	}
  }
  return a[:n]
}
func read_dirs(dir string) (dirs, files [2]map[string]os.FileInfo) {
  for i := 0; i < 2; i++ {
    // TODO inspect error -> retry if suffix "too many open files"?
    children, e := ioutil.ReadDir(filepath.Join(prefixes[i],dir))
    if e!=nil { fmt.Println("Error reading directory:",e) }
    // NOTE optional filter hidden files
    //children = filter(children, func(f os.FileInfo)bool{return f.Name()[0]!='.'})
    dirs[i],files[i] = children2map(children)
  }
  return
}

type FilePair struct { f0 os.FileInfo; f1 os.FileInfo }
func compare(m0, m1 map[string]os.FileInfo) (mboth map[string]FilePair, mA, mB map[string]os.FileInfo)  {
  // initialize
  mboth = make(map[string]FilePair)
  mA = make(map[string]os.FileInfo)
  mB = make(map[string]os.FileInfo)
  // compare
  for name, f0 := range m0 {
    if f1, ok := m1[name]; ok {
      mboth[name] = FilePair{f0, f1}
      delete(m1, name)
    } else {
      mA[name] = f0
    }
    delete(m0, name)
  }
  mB = m1
  //fmt.Println("both",mboth,"A",mA,"B",mB)
  return
}

func process(sem chan bool, dir string, wg *sync.WaitGroup) {
  // Decreasing internal counter for wait-group as soon as goroutine finishes
  defer wg.Done()

  // Sending increments, reading decrements semaphore
  sem <- true
  defer func() { <-sem }()

  //fmt.Println("Process item", dir)
  dirs, files := read_dirs(dir)
  dirsboth, dirsA, dirsB := compare(dirs[0], dirs[1])
  filesboth, filesA, filesB := compare(files[0], files[1])
  if len(dirsA)>0 || len(filesA)>0 || len(dirsB)>0 || len(filesB)>0 {
    // TODO format better via fmt.Printf()
    fmt.Println("<>",dir,":: OnlyA:",dirsA,filesA,"OnlyB:",dirsB,filesB)
  }

  for dn, _ := range dirsboth {
    wg.Add(1)
    go process( sem, filepath.Join( dir, dn ), wg )
  }
  for fn, f := range filesboth {
    // TODO compare content if wanted
    if f.f0.Size()!=f.f1.Size() || f.f0.ModTime()!=f.f1.ModTime() {
      fmt.Println("<>",dir,":: Files not equal:",fn)
    }
  }
}

func main() {
  wg := new(sync.WaitGroup)

  // initialize folders to compare
  if len(os.Args) == 3 {
    prefixes[0] = os.Args[1]
    prefixes[1] = os.Args[2]
  } else {
    return
  }

  // Create semaphore to limit number of workers
  // Without this limit, OS runs out of file descriptors and thus returns invalid results
  sem := make(chan bool, 100)

  // Adding root item
  wg.Add(1)
  go process(sem,".", wg)

  // Waiting for all goroutines to finish (otherwise they die as main routine dies)
  wg.Wait()
}
