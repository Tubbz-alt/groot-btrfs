// Code generated by counterfeiter. DO NOT EDIT.
package grootfakes

import (
	"sync"

	"code.cloudfoundry.org/lager"
	"github.com/SUSE/groot-btrfs/groot"
)

type FakeGarbageCollector struct {
	UnusedVolumesStub        func(logger lager.Logger, chainIDsToPreserve []string) ([]string, error)
	unusedVolumesMutex       sync.RWMutex
	unusedVolumesArgsForCall []struct {
		logger             lager.Logger
		chainIDsToPreserve []string
	}
	unusedVolumesReturns struct {
		result1 []string
		result2 error
	}
	unusedVolumesReturnsOnCall map[int]struct {
		result1 []string
		result2 error
	}
	MarkUnusedStub        func(logger lager.Logger, unusedVolumes []string) error
	markUnusedMutex       sync.RWMutex
	markUnusedArgsForCall []struct {
		logger        lager.Logger
		unusedVolumes []string
	}
	markUnusedReturns struct {
		result1 error
	}
	markUnusedReturnsOnCall map[int]struct {
		result1 error
	}
	CollectStub        func(logger lager.Logger) error
	collectMutex       sync.RWMutex
	collectArgsForCall []struct {
		logger lager.Logger
	}
	collectReturns struct {
		result1 error
	}
	collectReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeGarbageCollector) UnusedVolumes(logger lager.Logger, chainIDsToPreserve []string) ([]string, error) {
	var chainIDsToPreserveCopy []string
	if chainIDsToPreserve != nil {
		chainIDsToPreserveCopy = make([]string, len(chainIDsToPreserve))
		copy(chainIDsToPreserveCopy, chainIDsToPreserve)
	}
	fake.unusedVolumesMutex.Lock()
	ret, specificReturn := fake.unusedVolumesReturnsOnCall[len(fake.unusedVolumesArgsForCall)]
	fake.unusedVolumesArgsForCall = append(fake.unusedVolumesArgsForCall, struct {
		logger             lager.Logger
		chainIDsToPreserve []string
	}{logger, chainIDsToPreserveCopy})
	fake.recordInvocation("UnusedVolumes", []interface{}{logger, chainIDsToPreserveCopy})
	fake.unusedVolumesMutex.Unlock()
	if fake.UnusedVolumesStub != nil {
		return fake.UnusedVolumesStub(logger, chainIDsToPreserve)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.unusedVolumesReturns.result1, fake.unusedVolumesReturns.result2
}

func (fake *FakeGarbageCollector) UnusedVolumesCallCount() int {
	fake.unusedVolumesMutex.RLock()
	defer fake.unusedVolumesMutex.RUnlock()
	return len(fake.unusedVolumesArgsForCall)
}

func (fake *FakeGarbageCollector) UnusedVolumesArgsForCall(i int) (lager.Logger, []string) {
	fake.unusedVolumesMutex.RLock()
	defer fake.unusedVolumesMutex.RUnlock()
	return fake.unusedVolumesArgsForCall[i].logger, fake.unusedVolumesArgsForCall[i].chainIDsToPreserve
}

func (fake *FakeGarbageCollector) UnusedVolumesReturns(result1 []string, result2 error) {
	fake.UnusedVolumesStub = nil
	fake.unusedVolumesReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeGarbageCollector) UnusedVolumesReturnsOnCall(i int, result1 []string, result2 error) {
	fake.UnusedVolumesStub = nil
	if fake.unusedVolumesReturnsOnCall == nil {
		fake.unusedVolumesReturnsOnCall = make(map[int]struct {
			result1 []string
			result2 error
		})
	}
	fake.unusedVolumesReturnsOnCall[i] = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeGarbageCollector) MarkUnused(logger lager.Logger, unusedVolumes []string) error {
	var unusedVolumesCopy []string
	if unusedVolumes != nil {
		unusedVolumesCopy = make([]string, len(unusedVolumes))
		copy(unusedVolumesCopy, unusedVolumes)
	}
	fake.markUnusedMutex.Lock()
	ret, specificReturn := fake.markUnusedReturnsOnCall[len(fake.markUnusedArgsForCall)]
	fake.markUnusedArgsForCall = append(fake.markUnusedArgsForCall, struct {
		logger        lager.Logger
		unusedVolumes []string
	}{logger, unusedVolumesCopy})
	fake.recordInvocation("MarkUnused", []interface{}{logger, unusedVolumesCopy})
	fake.markUnusedMutex.Unlock()
	if fake.MarkUnusedStub != nil {
		return fake.MarkUnusedStub(logger, unusedVolumes)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.markUnusedReturns.result1
}

func (fake *FakeGarbageCollector) MarkUnusedCallCount() int {
	fake.markUnusedMutex.RLock()
	defer fake.markUnusedMutex.RUnlock()
	return len(fake.markUnusedArgsForCall)
}

func (fake *FakeGarbageCollector) MarkUnusedArgsForCall(i int) (lager.Logger, []string) {
	fake.markUnusedMutex.RLock()
	defer fake.markUnusedMutex.RUnlock()
	return fake.markUnusedArgsForCall[i].logger, fake.markUnusedArgsForCall[i].unusedVolumes
}

func (fake *FakeGarbageCollector) MarkUnusedReturns(result1 error) {
	fake.MarkUnusedStub = nil
	fake.markUnusedReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeGarbageCollector) MarkUnusedReturnsOnCall(i int, result1 error) {
	fake.MarkUnusedStub = nil
	if fake.markUnusedReturnsOnCall == nil {
		fake.markUnusedReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.markUnusedReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeGarbageCollector) Collect(logger lager.Logger) error {
	fake.collectMutex.Lock()
	ret, specificReturn := fake.collectReturnsOnCall[len(fake.collectArgsForCall)]
	fake.collectArgsForCall = append(fake.collectArgsForCall, struct {
		logger lager.Logger
	}{logger})
	fake.recordInvocation("Collect", []interface{}{logger})
	fake.collectMutex.Unlock()
	if fake.CollectStub != nil {
		return fake.CollectStub(logger)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.collectReturns.result1
}

func (fake *FakeGarbageCollector) CollectCallCount() int {
	fake.collectMutex.RLock()
	defer fake.collectMutex.RUnlock()
	return len(fake.collectArgsForCall)
}

func (fake *FakeGarbageCollector) CollectArgsForCall(i int) lager.Logger {
	fake.collectMutex.RLock()
	defer fake.collectMutex.RUnlock()
	return fake.collectArgsForCall[i].logger
}

func (fake *FakeGarbageCollector) CollectReturns(result1 error) {
	fake.CollectStub = nil
	fake.collectReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeGarbageCollector) CollectReturnsOnCall(i int, result1 error) {
	fake.CollectStub = nil
	if fake.collectReturnsOnCall == nil {
		fake.collectReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.collectReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeGarbageCollector) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.unusedVolumesMutex.RLock()
	defer fake.unusedVolumesMutex.RUnlock()
	fake.markUnusedMutex.RLock()
	defer fake.markUnusedMutex.RUnlock()
	fake.collectMutex.RLock()
	defer fake.collectMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeGarbageCollector) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ groot.GarbageCollector = new(FakeGarbageCollector)
