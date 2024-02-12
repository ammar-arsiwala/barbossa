package main

import "github.com/aka-mj/go-semaphore"

type Semaphore struct {
	sem semaphore.Semaphore
}

func OpenSem() (*Semaphore, error) {
	sem := semaphore.Semaphore{}
	if err := sem.Open(semaphore_name, 0644, 0); err != nil {
		if err := sem.Unlink(); err != nil {
			return nil, err
		}
		if err := sem.Open(semaphore_name, 0644, 0); err != nil {
			return nil, err
		}
	}

	return &Semaphore{
		sem: sem,
	}, nil
}

func (s *Semaphore) Signal() error {
	return s.sem.Post()
}

func WaitForParent() error {
	sem, err := OpenSem()
	if err != nil {
		return err
	}

	return sem.sem.Wait()
}

const semaphore_name = "BARB_SEMAPHORE"
