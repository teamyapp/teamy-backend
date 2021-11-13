package idgen

import (
	"github.com/stretchr/testify/assert"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/repo/repo_test"
	"testing"
)

func TestNextUniqueID(t *testing.T) {
	testCases := []struct {
		name string
		resourceType string
		expected oneEntity.ID
		expectedHasErr bool
	}{
		{
			name: "",
			resourceType: "",
			expected: "",
			expectedHasErr: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			idGenerator := newIDGenerator(repo_test.NewFakeIDRange(), 5)

			actual, err := idGenerator.NextUniqueID(testCase.resourceType)
			if testCase.expectedHasErr {
				assert.NotNil(t, err)
				return
			}
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

