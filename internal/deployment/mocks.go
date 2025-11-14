package deployment

import (
	"github.com/stretchr/testify/mock"
)

// MockCommandExecutor is a mock implementation of CommandExecutor
type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) Execute(command string) (*ExecutionResult, error) {
	args := m.Called(command)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExecutionResult), args.Error(1)
}

// MockComposerReader is a mock implementation of ComposerReader
type MockComposerReader struct {
	mock.Mock
}

func (m *MockComposerReader) ReadComposerData(filepath string) (map[string]interface{}, error) {
	args := m.Called(filepath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// MockTelemetryClient is a mock implementation of TelemetryClient
type MockTelemetryClient struct {
	mock.Mock
}

func (m *MockTelemetryClient) SendAndParseResponse(payload *Payload) (map[string]interface{}, error) {
	args := m.Called(payload)
	var response map[string]interface{}
	if args.Get(0) != nil {
		response = args.Get(0).(map[string]interface{})
	}
	return response, args.Error(1)
}
