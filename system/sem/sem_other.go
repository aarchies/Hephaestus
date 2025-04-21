package sem

// //#include <semaphore.h>      // For sem_*
// import "C"

// type Semaphore C.sem_t

// func New(value uint) (*Semaphore, error) {
// 	sem := new(Semaphore)

// 	if err := sem.Init(value); err != nil {
// 		return nil, err
// 	}

// 	return sem, nil
// }

// func (sem *Semaphore) Wait() error {
// 	_, err := C.sem_wait((*C.sem_t)(sem))
// 	return err
// }

// func (sem *Semaphore) TryWait() error {
// 	_, err := C.sem_trywait((*C.sem_t)(sem))
// 	return err
// }

// func (sem *Semaphore) Post() error {
// 	_, err := C.sem_post((*C.sem_t)(sem))
// 	return err
// }

// func (sem *Semaphore) Init(value uint) error {
// 	_, err := C.sem_init((*C.sem_t)(sem), 1, C.uint(value))
// 	return err
// }

// func (sem *Semaphore) Destroy() error {
// 	_, err := C.sem_destroy((*C.sem_t)(sem))
// 	return err
// }
