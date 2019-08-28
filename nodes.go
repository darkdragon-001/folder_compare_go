package main

import (
  "os"
  "path/filepath"
)


var prefixes [2]string
func SetPrefixes(A, B string) {
  prefixes[0] = A
  prefixes[1] = B
}


// FILE

type File struct {
  name_                string
  path_                string
  parent_              *Dir
  files                [2]os.FileInfo
  state_               State
}

// constructors
// TODO check fileA|B nil or .IsDir() -> read child string from there!
func NewFile(parent *Dir, name string, fileA, fileB os.FileInfo, state State) *File {
  path := name
  if parent != nil {
    path = filepath.Join(parent.path(), name)
  }
  return &File {
    name_: name,
    path_: path,
    parent_: parent,
    files: [2]os.FileInfo{fileA,fileB},
    state_: state,
  }
}
func (file File) New(parent *Dir, name string, fileA, fileB os.FileInfo, state State) *File {
  return NewFile(parent, name, fileA, fileB, state)
}

// getters/setters
func (file *File) name() string {
  return file.name_
}
func (file *File) path() string {
  return file.path_
}
func (file *File) paths() [2]string {
  return [2]string {
    filepath.Join(prefixes[0], file.path_),
    filepath.Join(prefixes[1], file.path_),
  }
}
func (file *File) parent() *Dir {
  return file.parent_
}
func (file *File) state() State {
  return file.state_
}
func (file *File) setState(state State) {
  file.state_ = state
}


// DIR

type Dir struct {
  info                 File
  dirs                 struct{single [2][]*Dir;both []*Dir}
  files                struct{single [2][]*File;both []*File}
}

// constructors
func NewDirRoot() *Dir {
  fileA, _ := os.Lstat(prefixes[0])
  fileB, _ := os.Lstat(prefixes[1])
  info := NewFile(nil, ".", fileA, fileB, STATE_UNKNOWN)
  return &Dir { info: *info }
}
func NewDir(parent *Dir, name string, fileA, fileB os.FileInfo, state State) *Dir {
  info := NewFile(parent, name, fileA, fileB, state)
  return &Dir { info: *info }
}
func (dir Dir) New(parent *Dir, name string, fileA, fileB os.FileInfo, state State) *Dir {
  return NewDir(parent, name, fileA, fileB, state)
}

// getters/setters
func (dir *Dir) name() string {
  return dir.info.name()
}
func (dir *Dir) path() string {
  return dir.info.path()
}
func (dir *Dir) paths() [2]string {
  return dir.info.paths()
}
func (dir *Dir) parent() *Dir {
  return dir.info.parent()
}
func (dir *Dir) setState(state State) {
  dir.info.setState(state)
}
func (dir *Dir) state() State {
  return dir.info.state()
}


// handle states
func (dir *Dir) isEqual() bool {
  // check if we calculated the result already
  if s := dir.state(); s != STATE_UNKNOWN {
    return state2bool(s)
  }
  // items only in A
  if len(dir.dirs.single[0])>0 || len(dir.files.single[0])>0 {
    return false
  }
  // items only in B
  if len(dir.dirs.single[1])>0 || len(dir.files.single[1])>0 {
    return false
  }
  // different file content
  for _, f := range dir.files.both {
    if !state2bool(f.state()) {
      return false
    }
  }
  // TODO recursive state of dir.dirs.both -> depth first search (while state==unknown update parent with equal, while state!=unequal update parent with unequal)
  return true
}
func (dir *Dir) updateStates() bool {
  // compute equality
  equal := true
  for _, n := range dir.dirs.both {
    e := n.updateStates()  // depth first search
    equal = equal && e
  }
  equal = equal && dir.isEqual()
  // update state
  if equal {
    for dir != nil && dir.state() == STATE_UNKNOWN {
      dir.setState(STATE_EQUAL)
      dir = dir.parent()
    }
  } else {
    for dir != nil && dir.state() != STATE_UNEQUAL {
      dir.setState(STATE_UNEQUAL)
      dir = dir.parent()
    }
  }
  // return
  return equal
}





