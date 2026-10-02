package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/tuliporg/agent-management-api/models"
)

type MockDatabaseHelper struct {
	*MockDatabase
}

func NewMockDatabaseHelper() *MockDatabaseHelper {
	return &MockDatabaseHelper{
		MockDatabase: new(MockDatabase),
	}
}

func (m *MockDatabaseHelper) ExpectFindAgentByEmail(email string, agent *models.Agent, err error) {
	m.On("Where", mock.MatchedBy(func(q interface{}) bool {
		return true
	}), mock.MatchedBy(func(args interface{}) bool {
		return true
	})).Return(m.MockDatabase)
	m.On("First", mock.MatchedBy(func(dest interface{}) bool {
		return true
	}), mock.Anything).Run(func(args mock.Arguments) {
		if agent != nil {
			destPtr := args.Get(0).(*models.Agent)
			*destPtr = *agent
		}
	}).Return(err)
}

func (m *MockDatabaseHelper) ExpectCreateAgent(agent *models.Agent, err error) {
	m.On("Create", mock.MatchedBy(func(value interface{}) bool {
		return true
	})).Return(err)
}

func (m *MockDatabaseHelper) ExpectUpdateAgent(agent *models.Agent, err error) {
	m.On("Save", mock.MatchedBy(func(value interface{}) bool {
		return true
	})).Return(err)
}

func (m *MockDatabaseHelper) ExpectDeleteAgent(id string, err error) {
	m.On("Where", mock.MatchedBy(func(q interface{}) bool {
		return true
	}), mock.Anything).Return(m.MockDatabase)
	m.On("Delete", mock.Anything, mock.Anything).Return(err)
}

func (m *MockDatabaseHelper) ExpectFindAllAgents(agents *[]models.Agent, err error) {
	m.On("Find", mock.MatchedBy(func(dest interface{}) bool {
		return true
	}), mock.Anything).Run(func(args mock.Arguments) {
		if agents != nil {
			destPtr := args.Get(0).(*[]models.Agent)
			*destPtr = *agents
		}
	}).Return(err)
}

func (m *MockDatabaseHelper) ExpectCountAgents(count int64, err error) {
	m.On("Model", mock.Anything).Return(m.MockDatabase)
	m.On("Count", mock.Anything).Run(func(args mock.Arguments) {
		countPtr := args.Get(0).(*int64)
		*countPtr = count
	}).Return(m.MockDatabase)
}
