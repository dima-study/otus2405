package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

// WithDone wraps current stage and returns wrapped one, which could be canceled.
// Close the done channel to cancel stage flow.
func (stage Stage) WithDone(done In) Stage {
	return func(in In) Out {
		// wrappedIn is a new input channel for original stage.
		// It will be closed once done is closed.
		wrappedIn := make(Bi)

		go func() {
			// If stage is canceled, there could be "blocked" senders.
			// Should we unblock them?
			defer func() {
				for {
					_, ok := <-in
					if !ok {
						return
					}
				}
			}()

			defer close(wrappedIn)

			// Read from original input while possible.
			for {
				// Priority cancel.
				select {
				case <-done:
					return
				default:
				}

				// Cancel or read from original input.
				select {
				case <-done:
					return
				case val, ok := <-in:
					// Original input also could be closed: catch it!
					if !ok {
						return
					}

					// Cancel or send.
					select {
					case <-done:
						return
					case wrappedIn <- val:
					}
				}
			}
		}()

		// Make the original stage work with wrapped input.
		return stage(wrappedIn)
	}
}

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	if len(stages) == 0 {
		return nil
	}

	pipeline := stages[0].WithDone(done)(in)
	for _, stage := range stages[1:] {
		pipeline = stage.WithDone(done)(pipeline)
	}

	return pipeline
}
