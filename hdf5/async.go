package hdf5

import (
	"errors"
	"sync"
)

type AsyncResult struct {
	data         interface{}
	err          error
	done         chan struct{}
	completeOnce sync.Once
}

func NewAsyncResult() *AsyncResult {
	return &AsyncResult{
		done: make(chan struct{}),
	}
}

func (r *AsyncResult) Wait() error {
	<-r.done
	return r.err
}

func (r *AsyncResult) GetData() interface{} {
	return r.data
}

func (r *AsyncResult) GetError() error {
	return r.err
}

func (r *AsyncResult) complete(data interface{}, err error) {
	r.completeOnce.Do(func() {
		r.data = data
		r.err = err
		close(r.done)
	})
}

type AsyncDataset interface {
	ReadAsync(data interface{}) *AsyncResult
	WriteAsync(data interface{}) *AsyncResult
	ReadSliceAsync(data interface{}, sel Selection) *AsyncResult
	WriteSliceAsync(data interface{}, sel Selection) *AsyncResult
}

func (d *dataset) ReadAsync(data interface{}) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		err := d.Read(data)
		result.complete(data, err)
	}()

	return result
}

func (d *dataset) WriteAsync(data interface{}) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		err := d.Write(data)
		result.complete(nil, err)
	}()

	return result
}

func (d *dataset) ReadSliceAsync(data interface{}, sel Selection) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		err := d.ReadSlice(data, sel)
		result.complete(data, err)
	}()

	return result
}

func (d *dataset) WriteSliceAsync(data interface{}, sel Selection) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		err := d.WriteSlice(data, sel)
		result.complete(nil, err)
	}()

	return result
}

type AsyncFile interface {
	OpenGroupAsync(name string) *AsyncResult
	CreateGroupAsync(name string) *AsyncResult
	OpenDatasetAsync(name string) *AsyncResult
	CreateDatasetAsync(name string, dtype Datatype, space Dataspace, plist PropertyList) *AsyncResult
}

func (f *File) OpenGroupAsync(name string) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		group, err := f.GetGroup(name)
		result.complete(group, err)
	}()

	return result
}

func (f *File) CreateGroupAsync(name string) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		group, err := f.CreateGroup(name)
		result.complete(group, err)
	}()

	return result
}

func (f *File) OpenDatasetAsync(name string) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		dset, err := f.GetDataset(name)
		result.complete(dset, err)
	}()

	return result
}

func (f *File) CreateDatasetAsync(name string, dtype Datatype, space Dataspace, plist PropertyList) *AsyncResult {
	result := NewAsyncResult()

	go func() {
		dset, err := f.CreateDataset(name, dtype, space, plist)
		result.complete(dset, err)
	}()

	return result
}

type AsyncManager struct {
	wg      sync.WaitGroup
	running bool
	mutex   sync.Mutex
}

func NewAsyncManager() *AsyncManager {
	return &AsyncManager{
		running: true,
	}
}

func (m *AsyncManager) Submit(f func() error) *AsyncResult {
	result := NewAsyncResult()

	m.mutex.Lock()
	if !m.running {
		m.mutex.Unlock()
		result.complete(nil, errors.New("hdf5: async manager is stopped"))
		return result
	}
	m.wg.Add(1)
	m.mutex.Unlock()

	go func() {
		defer m.wg.Done()
		err := f()
		result.complete(nil, err)
	}()

	return result
}

func (m *AsyncManager) Wait() {
	m.wg.Wait()
}

func (m *AsyncManager) Stop() {
	m.mutex.Lock()
	m.running = false
	m.mutex.Unlock()
	m.wg.Wait()
}
