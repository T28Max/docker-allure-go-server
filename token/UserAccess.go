// Copyright 2024 Maxim Tverdohleb <tverdohleb.maxim@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package token

import "allure-server/utils"

type UserAccess struct {
	UserName string
	Roles    []string
}

func (ua *UserAccess) GetRoles() []string {
	return ua.Roles
}

func (ua *UserAccess) GetUsername() string {
	return ua.UserName
}
func (ua *UserAccess) CheckAccess(role string) bool {
	if role == "" {
		return false
	}
	return utils.StringInArray(role, ua.Roles)
}
