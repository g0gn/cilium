// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package cmd

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/envoy-config-input.json
	envoyConfigInput string

	//go:embed testdata/envoy-config-output.json
	envoyConfigOutput string
)

func Test_jsonFieldMaskPostProcess(t *testing.T) {
	type args struct {
		input      []byte
		fieldNames []string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "simple struct",
			args: args{
				input: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com"
				}`),
				fieldNames: []string{"password"},
			},
			want: []byte(`{
				"username": "user1",
				"password": "[redacted]",
				"email": "user1@example.com"
			}`),
		},
		{
			name: "array struct",
			args: args{
				input: []byte(`{
					"username": "user1",
					"secrets": [
						{
							"password": "mypassword"
						},
						{
							"password": "anotherone"
						}
					],
					"password": "mypassword",
					"email": "user1@example.com"
				}`),
				fieldNames: []string{"password"},
			},
			want: []byte(`{
					"username": "user1",
					"secrets": [
						{
							"password": "[redacted]"
						},
						{
							"password": "[redacted]"
						}
					],
					"password": "[redacted]",
					"email": "user1@example.com"
			}`),
		},
		{
			name: "nested struct",
			args: args{
				input: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com",
					"complex": {
						"password": "anotherpassword"
					}
				}`),
				fieldNames: []string{"password"},
			},
			want: []byte(`{
				"username": "user1",
				"password": "[redacted]",
				"email": "user1@example.com",
				"complex": {
					"password": "[redacted]"
				}
			}`),
		},
		{
			name: "nested array struct",
			args: args{
				input: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com",
					"complex": {
						"secrets": [
							{
								"password": "mypassword"
							},
							{
								"password": "anotherpassword"
							}
						]
					}
				}`),
				fieldNames: []string{"password"},
			},
			want: []byte(`{
				"username": "user1",
				"password": "[redacted]",
				"email": "user1@example.com",
				"complex": {
					"secrets": [
						{
							"password": "[redacted]"
						},
						{
							"password": "[redacted]"
						}
					]
				}
			}`),
		},
		{
			name: "no masked field",
			args: args{
				input: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com",
					"complex": {
						"password": "anotherpassword"
					}
				}`),
				fieldNames: []string{"no-such-field"},
			},
			want: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com",
					"complex": {
						"password": "anotherpassword"
					}
				}`),
		},
		{
			name: "mask object field",
			args: args{
				input: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com",
					"complex": {
						"password": "anotherpassword"
					}
				}`),
				fieldNames: []string{"complex"},
			},
			want: []byte(`{
					"username": "user1",
					"password": "mypassword",
					"email": "user1@example.com",
					"complex": "[redacted]"
				}`),
		},
		{
			name: "invalid input",
			args: args{
				input:      []byte(`{"username": "user1",}`),
				fieldNames: []string{"password"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jsonFieldMaskPostProcess(tt.args.fieldNames)(tt.args.input)
			require.Equal(t, tt.wantErr, err != nil)
			// only assert the output if there is no error
			// as JSONEq func is used to compare the output
			if !tt.wantErr {
				require.JSONEq(t, string(tt.want), string(got))
			}
		})
	}
}

// Test_jsonFieldMaskPostProcessFile is to make sure that the json field masking
// is working correctly with the actual sample envoy config file
func Test_jsonFieldMaskPostProcessFile(t *testing.T) {
	res, err := jsonFieldMaskPostProcess([]string{
		"api_key",
		"trusted_ca",
		"certificate_chain",
		"private_key",
	})([]byte(envoyConfigInput))

	require.NoError(t, err)
	require.JSONEq(t, envoyConfigOutput, string(res))
}
