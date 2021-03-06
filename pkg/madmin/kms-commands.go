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

package madmin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// CreateKey tries to create a new master key with the given keyID
// at the KMS connected to a MinIO server.
func (adm *AdminClient) CreateKey(ctx context.Context, keyID string) error {
	// POST /minio/admin/v3/kms/key/create?key-id=<keyID>
	qv := url.Values{}
	qv.Set("key-id", keyID)
	reqData := requestData{
		relPath:     adminAPIPrefix + "/kms/key/create",
		queryValues: qv,
	}

	resp, err := adm.executeMethod(ctx, http.MethodPost, reqData)
	if err != nil {
		return err
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return httpRespToErrorResponse(resp)
	}
	return nil
}

// GetKeyStatus requests status information about the key referenced by keyID
// from the KMS connected to a MinIO by performing a Admin-API request.
// It basically hits the `/minio/admin/v3/kms/key/status` API endpoint.
func (adm *AdminClient) GetKeyStatus(ctx context.Context, keyID string) (*KMSKeyStatus, error) {
	// GET /minio/admin/v3/kms/key/status?key-id=<keyID>
	qv := url.Values{}
	qv.Set("key-id", keyID)
	reqData := requestData{
		relPath:     adminAPIPrefix + "/kms/key/status",
		queryValues: qv,
	}

	resp, err := adm.executeMethod(ctx, http.MethodGet, reqData)
	if err != nil {
		return nil, err
	}
	defer closeResponse(resp)
	if resp.StatusCode != http.StatusOK {
		return nil, httpRespToErrorResponse(resp)
	}
	var keyInfo KMSKeyStatus
	if err = json.NewDecoder(resp.Body).Decode(&keyInfo); err != nil {
		return nil, err
	}
	return &keyInfo, nil
}

// KMSKeyStatus contains some status information about a KMS master key.
// The MinIO server tries to access the KMS and perform encryption and
// decryption operations. If the MinIO server can access the KMS and
// all master key operations succeed it returns a status containing only
// the master key ID but no error.
type KMSKeyStatus struct {
	KeyID         string `json:"key-id"`
	EncryptionErr string `json:"encryption-error,omitempty"` // An empty error == success
	DecryptionErr string `json:"decryption-error,omitempty"` // An empty error == success
}
