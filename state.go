package main


// state enum
type State int
const (
  STATE_UNKNOWN State = iota
  STATE_EQUAL
  STATE_UNEQUAL
  STATE_A_ONLY
  STATE_B_ONLY
)


func bool2state(b bool) State {
  if b {
    return STATE_EQUAL
  } else {
    return STATE_UNEQUAL
  }
}
func state2bool(s State) bool {
  return ! ( s != STATE_EQUAL )
}

func state2byte(s State) byte {
  if s == STATE_EQUAL {
    return '='
  } else if s == STATE_UNEQUAL {
    return '*'
  } else if s == STATE_A_ONLY {
    return 'A'
  } else if s == STATE_B_ONLY {
    return 'B'
  } else {
    return '?'
  }
}
