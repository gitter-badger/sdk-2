// Copyright 2017 Stratumn SAS. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tmpoptestcasesmocks

import (
	"github.com/stratumn/sdk/tmpop"
	"github.com/stretchr/testify/mock"
)

// MockedTendermintClient is a mock for the tmpop.TendermintClient interface
type MockedTendermintClient struct {
	mock.Mock
}

// AllowCalls allows any call to go through the mock without throwing errors
func (m *MockedTendermintClient) AllowCalls() {
	m.On("Block", mock.Anything)
}

// Block returns an empty block
func (m *MockedTendermintClient) Block(height int) *tmpop.Block {
	args := m.Called(height)
	return args.Get(0).(*tmpop.Block)
}
