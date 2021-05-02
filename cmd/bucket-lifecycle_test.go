// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"net/http"
	"testing"
	"time"

	xhttp "github.com/minio/minio/cmd/http"
)

// TestParseRestoreObjStatus tests parseRestoreObjStatus
func TestParseRestoreObjStatus(t *testing.T) {
	testCases := []struct {
		restoreHdr     string
		expectedStatus restoreObjStatus
		expectedErr    error
	}{
		{
			// valid: represents a restored object, 'pending' expiry.
			restoreHdr: "ongoing-request=false, expiry-date=Fri, 21 Dec 2012 00:00:00 GMT",
			expectedStatus: restoreObjStatus{
				ongoing: false,
				expiry:  time.Date(2012, 12, 21, 0, 0, 0, 0, time.UTC),
			},
			expectedErr: nil,
		},
		{
			// valid: represents an ongoing restore object request.
			restoreHdr: "ongoing-request=true",
			expectedStatus: restoreObjStatus{
				ongoing: true,
			},
			expectedErr: nil,
		},
		{
			// invalid; ongoing restore object request can't have expiry set on it.
			restoreHdr:     "ongoing-request=true, expiry-date=Fri, 21 Dec 2012 00:00:00 GMT",
			expectedStatus: restoreObjStatus{},
			expectedErr:    errRestoreHDRMalformed,
		},
		{
			// invalid; completed restore object request must have expiry set on it.
			restoreHdr:     "ongoing-request=false",
			expectedStatus: restoreObjStatus{},
			expectedErr:    errRestoreHDRMalformed,
		},
	}
	for i, tc := range testCases {
		actual, err := parseRestoreObjStatus(tc.restoreHdr)
		if err != tc.expectedErr {
			t.Fatalf("Test %d: got %v expected %v", i+1, err, tc.expectedErr)
		}
		if actual != tc.expectedStatus {
			t.Fatalf("Test %d: got %v expected %v", i+1, actual, tc.expectedStatus)
		}
	}
}

// TestRestoreObjStatusRoundTrip restoreObjStatus roundtrip
func TestRestoreObjStatusRoundTrip(t *testing.T) {
	testCases := []restoreObjStatus{
		ongoingRestoreObj(),
		completedRestoreObj(time.Now().UTC()),
	}
	for i, tc := range testCases {
		actual, err := parseRestoreObjStatus(tc.String())
		if err != nil {
			t.Fatalf("Test %d: parse restore object failed: %v", i+1, err)
		}
		if actual.ongoing != tc.ongoing || actual.expiry.Format(http.TimeFormat) != tc.expiry.Format(http.TimeFormat) {
			t.Fatalf("Test %d: got %v expected %v", i+1, actual, tc)
		}
	}
}

// TestRestoreObjOnDisk tests restoreObjStatus' OnDisk method
func TestRestoreObjOnDisk(t *testing.T) {
	testCases := []struct {
		restoreStatus restoreObjStatus
		ondisk        bool
	}{
		{
			// restore in progress
			restoreStatus: ongoingRestoreObj(),
			ondisk:        false,
		},
		{
			// restore completed but expired
			restoreStatus: completedRestoreObj(time.Now().Add(-time.Hour)),
			ondisk:        false,
		},
		{
			// restore completed
			restoreStatus: completedRestoreObj(time.Now().Add(time.Hour)),
			ondisk:        true,
		},
	}

	for i, tc := range testCases {
		if actual := tc.restoreStatus.OnDisk(); actual != tc.ondisk {
			t.Fatalf("Test %d: expected %v but got %v", i+1, tc.ondisk, actual)
		}
	}
}

// TestIsRestoredObjectOnDisk tests isRestoredObjectOnDisk helper function
func TestIsRestoredObjectOnDisk(t *testing.T) {
	testCases := []struct {
		meta   map[string]string
		ondisk bool
	}{
		{
			// restore in progress
			meta: map[string]string{
				xhttp.AmzRestore: ongoingRestoreObj().String(),
			},
			ondisk: false,
		},
		{
			// restore completed
			meta: map[string]string{
				xhttp.AmzRestore: completedRestoreObj(time.Now().Add(time.Hour)).String(),
			},
			ondisk: true,
		},
		{
			// restore completed but expired
			meta: map[string]string{
				xhttp.AmzRestore: completedRestoreObj(time.Now().Add(-time.Hour)).String(),
			},
			ondisk: false,
		},
	}

	for i, tc := range testCases {
		if actual := isRestoredObjectOnDisk(tc.meta); actual != tc.ondisk {
			t.Fatalf("Test %d: expected %v but got %v for %v", i+1, tc.ondisk, actual, tc.meta)
		}
	}
}
