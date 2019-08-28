package main

import (
  "fmt"
)


// FILE

func (file *File) str() string {
  return fmt.Sprintf("[%s]%s", string(state2byte(file.state_)), file.path())
}


// DIR


// inspect contents of dir struct
// TODO template functions!?
func strFormatDirs(nodes []*Dir) string {
  var str string
  for _, n := range nodes {
    str += (*n).name()
    if !state2bool( (*n).state() ) {
      str += "*"
    }
    str += ", "  // TODO not for last item
  }
  return str
}
func strFormatFiles(nodes []*File) string {
  var str string
  for _, n := range nodes {
    str += (*n).name()
    if !state2bool( (*n).state() ) {
      str += "*"
    }
    str += ", "  // TODO not for last item
  }
  return str
}
func (dir *Dir) str() string {
  str := fmt.Sprintf( "%s:\n", dir.path() )
  str += fmt.Sprintf( "%4s: { %s } { %s }\n", "=", strFormatDirs(dir.dirs.both),      strFormatFiles(dir.files.both) )
  str += fmt.Sprintf( "%4s: { %s } { %s }\n", "A", strFormatDirs(dir.dirs.single[0]), strFormatFiles(dir.files.single[0]) )
  str += fmt.Sprintf( "%4s: { %s } { %s }"  , "B", strFormatDirs(dir.dirs.single[1]), strFormatFiles(dir.files.single[1]) )
  return str
}


// format changes: one node per line
// alternatives: dir.name() dir.path() dir.str()
func formatChangesDirs(dirs []*Dir) string {
  s := ""
  for _, n := range dirs {
    s += fmt.Sprintf("[%s]%s\n", string(state2byte(n.state())), n.path())
  }
  return s
}
func formatChangesFiles(files []*File) string {
  s := ""
  for _, n := range files {
    s += fmt.Sprintf("[%s]%s\n", string(state2byte(n.state())), n.path())
  }
  return s
}
func (dir *Dir) formatChanges() string {
  // init
  equ := true
  str := ""
  // single
  for i := 0; i < 2; i++ {
    s := formatChangesDirs( dir. dirs.single[i]); equ = equ && len(s)<1; str += s
    s  = formatChangesFiles(dir.files.single[i]); equ = equ && len(s)<1; str += s
  }
  // both
  for _, n := range dir.files.both {
    if !state2bool(n.state()) {
      equ = false
      str += fmt.Sprintf("[*]%s\n", n.path())
    }
  }
  // print
  var output string;
  //output += fmt.Sprintf("[%s]%s\n",string(state2byte(dir.state())),dir.path())  // NOTE dir.updateStates() should be called before
  if !equ {
    //output += fmt.Sprintf("[*]%s\n", dir.path())  // print folder name when children unequal
    output += fmt.Sprintf("%s", str)               // print unequal children
  }
  return output
}
func (dir *Dir) printRecursive() {
  // print all dirs in both
  fmt.Print(dir.formatChanges())
  // recurse
  for _, d := range dir.dirs.both {
    d.printRecursive()
  }
}




