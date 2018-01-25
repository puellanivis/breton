// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package user allows user account lookups by name or id.
package user

import (
	"os/user"
)

var (
	userImplemented  = true // set to false by lookup_stubs.go's init
	groupImplemented = true // set to false by lookup_stubs.go's init
)

// User represents a user account.
type User = user.User

// Group represents a grouping of users.
//
// On POSIX systems Gid contains a decimal number representing the group ID.
type Group = user.Group

// UnknownUserIdError is returned by LookupId when a user cannot be found.
type UnknownUserIdError = user.UnknownUserIdError

// UnknownUserError is returned by Lookup when
// a user cannot be found.
type UnknownUserError = user.UnknownUserError

// UnknownGroupIdError is returned by LookupGroupId when
// a group cannot be found.
type UnknownGroupIdError = user.UnknownGroupIdError

// UnknownGroupError is returned by LookupGroup when
// a group cannot be found.
type UnknownGroupError = user.UnknownGroupError
