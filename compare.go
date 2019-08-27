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

// TODO is is possible to create some sort of template so declare function only once?
// TODO create compare function which works without converting first to map
func compareDirs(dir *Dir, m0, m1 map[string]os.FileInfo) *Dir {
  var NodeFactory Dir
  nodes := &dir.dirs
  for name, f0 := range m0 {
    if f1, ok := m1[name]; ok {
      nodes.both = append(nodes.both, NodeFactory.New(dir, name, f0, f1, STATE_UNKNOWN))  // TODO only equal when all childs equal; NOTE comparison of TimeModified possible
      delete(m1, name)
    } else {
      nodes.single[0] = append(nodes.single[0], NodeFactory.New(dir, name, f0, nil, STATE_A_ONLY))
    }
    delete(m0, name)
  }
  nodes.single[1] = make([]*Dir, 0, len(m1))  // allocate space
  for name, f := range m1 {
    nodes.single[1] = append(nodes.single[1], NodeFactory.New(dir, name, nil, f, STATE_B_ONLY))
  }
  return dir
}
func fileInfoEqual(dir *Dir, f0, f1 os.FileInfo) bool {
  if f0.Size() != f1.Size() { return false }
  //if f0.ModTime() != f1.ModTime() { return false }
  return true
}
func fileContentEqual(dir *Dir, f0s, f1s os.FileInfo) bool {
  // compare file size
  if f0s.Size() != f1s.Size() {
    return false
  }

  // open files for reading
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
func compareFiles(dir *Dir, m0, m1 map[string]os.FileInfo) *Dir {
  var NodeFactory File
  nodes := &dir.files
  for name, f0 := range m0 {
    if f1, ok := m1[name]; ok {
      //filesEqual := fileInfoEqual(dir, f0,f1)
      filesEqual := fileContentEqual(dir, f0,f1)  // TODO check type -> sockets etc. can't be opened for reading, broken/dangling symlink, ...
      nodes.both = append(nodes.both, NodeFactory.New(dir, name, f0, f1, bool2state(filesEqual)))
      delete(m1, name)
    } else {
      nodes.single[0] = append(nodes.single[0], NodeFactory.New(dir, name, f0, nil, STATE_A_ONLY))
    }
    delete(m0, name)
  }
  nodes.single[1] = make([]*File, 0, len(m1))  // allocate space
  for name, f := range m1 {
    nodes.single[1] = append(nodes.single[1], NodeFactory.New(dir, name, nil, f, STATE_B_ONLY))
  }
  return dir
}


// PROCESS DIRECTORIES

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
func read_dirs(dir *Dir) (dirs, files [2]map[string]os.FileInfo) {
  paths := dir.paths()
  for i := 0; i < 2; i++ {
    // TODO inspect error -> retry if suffix "too many open files"?
    children, e := ioutil.ReadDir(paths[i])
    if e!=nil { fmt.Println("Error reading directory:",e) }
    // NOTE optional filter hidden files
    //children = filter(children, func(f os.FileInfo)bool{return f.Name()[0]!='.'})
    dirs[i],files[i] = children2map(children)
  }
  return
}

func processDir(dir *Dir) {
  //fmt.Println("Process", dir.path())
  dirs, files := read_dirs(dir)
  compareDirs(dir, dirs[0], dirs[1])
  compareFiles(dir, files[0], files[1])
}

