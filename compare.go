package main

import (
  "fmt"
  "os"
  "io"
  "io/ioutil"
  "bytes"
  "log"
  "path/filepath"
)


// COMPARING NODES

func fileInfoEqual(dir *Dir, f0, f1 os.FileInfo) bool {
  if f0.Size() != f1.Size() { return false }
  //if f0.ModTime() != f1.ModTime() { return false }
  return true
}
// source: https://stackoverflow.com/a/30038571/3779655
func fileContentEqual(dir *Dir, f0s, f1s os.FileInfo) bool {
  // compare file size
  if f0s.Size() != f1s.Size() {
    return false
  }

  // open files for reading
  // TODO check type -> sockets etc. can't be opened for reading, broken/dangling symlink, ...
  paths := dir.paths()
  f0, err := os.Open(filepath.Join(paths[0],f0s.Name()))
  defer f0.Close()
  if err != nil {
    log.Print(err)
    return false  // TODO better throw/catch exception and keep STATE_UNKNOWN
  }
  f1, err := os.Open(filepath.Join(paths[1],f0s.Name()))
  defer f1.Close()
  if err != nil {
    log.Print(err)
    return false  // TODO better throw/catch exception and keep STATE_UNKNOWN
  }

  // compare
  const chunckSize = 64000
  for {
    b1 := make([]byte, chunckSize)
    _, err1 := f0.Read(b1)

    b2 := make([]byte, chunckSize)
    _, err2 := f1.Read(b2)

    if err1 != nil || err2 != nil {
      if err1 == io.EOF && err2 == io.EOF {
        return true
      } else if err1 == io.EOF && err2 == io.EOF {
        return false
      } else {
        log.Print(err1, err2)
        return false  // TODO better throw/catch exception and keep STATE_UNKNOWN
      }
    }

    if !bytes.Equal(b1, b2) {
      return false
    }
  }
}

// NOTE slice sorted -> this algorithm, unsorted -> maps: https://stackoverflow.com/a/23874300/3779655
// TODO use additional data type for single which is same for dirs/files
func compareNodes(dir *Dir, A, B []os.FileInfo) *Dir {
  var i, j int
  for i < len(A) && j < len(B) {
    if A[i].Name() < B[j].Name() {
      // A only
      name := A[i].Name()
      if A[i].IsDir() {
        dir.dirs.single[0] = append(dir.dirs.single[0], NewDir(dir, name, A[i], nil, STATE_A_ONLY))
      } else {
        dir.files.single[0] = append(dir.files.single[0], NewFile(dir, name, A[i], nil, STATE_A_ONLY))
      }
      i++
    } else if A[i].Name() > B[j].Name() {
      // B only
      name := B[j].Name()
      if A[i].IsDir() {
        dir.dirs.single[1] = append(dir.dirs.single[1], NewDir(dir, name, nil, B[j], STATE_A_ONLY))
      } else {
        dir.files.single[1] = append(dir.files.single[1], NewFile(dir, name, nil, B[j], STATE_A_ONLY))
      }
      j++
    } else {
      // both
      name := A[i].Name()
      if A[i].IsDir() && B[j].IsDir() {
        dir.dirs.both = append(dir.dirs.both, NewDir(dir, name, A[i], B[j], STATE_UNKNOWN))
      } else if !A[i].IsDir() && !B[j].IsDir() {
        // compare files
        filesEqual := fileInfoEqual(dir, A[i], B[j])
        //filesEqual := fileContentEqual(dir, A[i], B[j])
        dir.files.both = append(dir.files.both, NewFile(dir, name, A[i], B[j], bool2state(filesEqual)))
      } else {
        if A[i].IsDir() {  // A dir, B file
          dir.dirs.single[0] = append(dir.dirs.single[0], NewDir(dir, name, A[i], nil, STATE_A_ONLY))
          dir.files.single[1] = append(dir.files.single[1], NewFile(dir, name, nil, B[j], STATE_A_ONLY))
        } else {  // A file, B dir
          dir.files.single[0] = append(dir.files.single[0], NewFile(dir, name, A[i], nil, STATE_A_ONLY))
          dir.dirs.single[1] = append(dir.dirs.single[1], NewDir(dir, name, nil, B[j], STATE_A_ONLY))
        }
      }
      i++
      j++
    }
  }
  return dir
}


// PROCESS DIRECTORIES

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

func processDir(dir *Dir) {
  //fmt.Println("Process", dir.path())
  paths := dir.paths()
  var nodes [2][]os.FileInfo; var e error
  for i := 0; i < 2; i++ {
    // TODO inspect error -> retry if suffix "too many open files"?
    nodes[i], e = ioutil.ReadDir(paths[i])
    if e!=nil { fmt.Println("Error reading directory:",e) }
    // NOTE optional filter hidden files
    //nodes[i] = filter(nodes[i], func(f os.FileInfo)bool{return f.Name()[0]!='.'})
  }
  compareNodes(dir, nodes[0], nodes[1])
}

