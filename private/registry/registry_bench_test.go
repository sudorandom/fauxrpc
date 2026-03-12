package registry

import (
	"fmt"
	"testing"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// A dummy service descriptor
type dummyServiceDescriptor struct {
    protoreflect.ServiceDescriptor
    name protoreflect.FullName
}

func (d dummyServiceDescriptor) FullName() protoreflect.FullName {
    return d.name
}

func BenchmarkForEachService_Current(b *testing.B) {
	r, err := NewServiceRegistry()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		r.addService(dummyServiceDescriptor{name: protoreflect.FullName(fmt.Sprintf("service_%d", i))})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.ForEachService(func(sd protoreflect.ServiceDescriptor) bool {
			return true
		})
	}
}

func BenchmarkForEachService_InsideRLock(b *testing.B) {
	r, err := NewServiceRegistry()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		r.addService(dummyServiceDescriptor{name: protoreflect.FullName(fmt.Sprintf("service_%d", i))})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.lock.RLock()
		for range r.services {
			if !true {
				break
			}
		}
		r.lock.RUnlock()
	}
}

func BenchmarkForEachService_Slice(b *testing.B) {
	r, err := NewServiceRegistry()
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < 1000; i++ {
		r.addService(dummyServiceDescriptor{name: protoreflect.FullName(fmt.Sprintf("service_%d", i))})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.lock.RLock()
		services := make([]protoreflect.ServiceDescriptor, 0, len(r.services))
		for _, s := range r.services {
			services = append(services, s)
		}
		r.lock.RUnlock()
		for range services {
			if !true {
				break
			}
		}
	}
}
