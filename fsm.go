package telgitbot

//Implementation of Finite-State Machine


type FSM struct {
	//states has a current registered states
	states  []string
	//store has a current state and available inputs
	store   map[string][]string
	//nextstates hash a transition to other states from current
	nextstates map[string][]string

	currentstate  string

}

//NewFSM provides construction of new FSM object
func NewFSM()*FSM {
	fsm := new(FSM)
	fsm.states = []string{}
	fsm.store = map[string][]string{}
	fsm.nextstates = map[string][]string{}
	return fsm
}

//AddState provides set new state
func (fsm *FSM) AddState(state string, nextstates []string, inp []string) {
	fsm.states = append(fsm.states, state)
	fsm.store[state] = inp
	fsm.nextstates[state] = nextstates

}

//SetState returns next available states
func (fsm *FSM) SetState(state string)[]string {
	next, ok := fsm.nextstates[state]
	if !ok {
		return []string{}
	}
	fsm.currentstate = state
	return next
}

//existNextState provides checking nextstate from current state
func (fsm *FSM) ExistNextState(nextstate string) bool {
	if !fsm.checkStates(fsm.currentstate) {
		return false
	}
	next, _ := fsm.nextstates[fsm.currentstate]
	for _, st := range next {
		if st == nextstate {
			return true
		}
	}

	return false
}

//CurrentState returns current state
func (fsm *FSM) CurrentState() string {
	return fsm.currentstate
}


func (fsm *FSM) checkStates(state string) bool {
	for _, inp := range fsm.states {
		if inp ==  state {
			return true
		}
	} 

	return false
}
