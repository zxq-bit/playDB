package repl

type Repl interface {
	Run(stopCh chan struct{}) error
	Step()
}
