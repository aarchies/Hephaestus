//go:generate sh -c "GOARCH=386 go tool cgo -godefs sem_linux.go | gofmt -r 'atomic.AddUintptr->atomic.AddUint32' | gofmt -r 'atomic.LoadUintptr->atomic.LoadUint32' | gofmt -r '(*uintptr)(unsafe.Pointer(x))->x' | gofmt -r '^uintptr(0)->^uint32(0)' | gofmt -r '(*uint32)(x)->x' | gofmt -r '(*uint64)(x)->x' > sem_linux_386.go"
//go:generate sh -c "GOARCH=amd64 go tool cgo -godefs sem_linux.go | gofmt -r 'atomic.AddUintptr->atomic.AddUint64' | gofmt -r 'atomic.LoadUintptr->atomic.LoadUint64' | gofmt -r '(*uintptr)(unsafe.Pointer(x))->x' | gofmt -r '^uintptr(0)->^uint64(0)' | gofmt -r '(*uint32)(x)->x' | gofmt -r '(*uint64)(x)->x' > sem_linux_amd64.go"

package sem
