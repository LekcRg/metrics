package crypto

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateHMAC(t *testing.T) {
	type args struct {
		content string
		key     string
	}
	type test struct {
		name string
		args args
		want string
	}

	const content = "Hello, world!"
	const key = "Sup3rP4swd"

	tests := []test{
		{
			name: "Text with key different",
			args: args{
				content: content,
				key:     key,
			},
			want: "d61e08ab978e60dc08ed0bedd1da80b65fcd42ca2948c1546624d49bbc1ddb04",
		},
		{
			name: "Text with key equal",
			args: args{
				content: content,
				key:     content,
			},
			want: "f662e3144e9f9b79f3bc7e926572cea9b50484875d8ee8531f5cdafbe79ecc09",
		},
		{
			name: "Empty text with key",
			args: args{
				content: "",
				key:     key,
			},
			want: "c411768e58628eb1e5d803db247ad18f68b107d8c5f6c133896df25488c78fb8",
		},
		{
			name: "Text without key",
			args: args{
				content: content,
				key:     "",
			},
			want: "0d192eb5bc5e4407192197cbf9e1658295fa3ff995b3ff914f3cc7c38d83b10f",
		},
		{
			name: "Empty content with empty key",
			args: args{
				content: "",
				key:     "",
			},
			want: "b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad",
		},
		{
			name: "Text is unicode with key",
			args: args{
				content: "Привет, 🐻!",
				key:     key,
			},
			want: "5becbf9b17148fa816e42e78620bb0529a36f6c78947205d80b2e6b148f79ef5",
		},
		{
			name: "Text len is 10000 with key",
			args: args{
				content: strings.Repeat("W", 10000),
				key:     key,
			},
			want: "2a5b6f75ead78ab43d18b94e1a48827f329ee8f466f7c753c19fa26b147c5444",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateHMAC([]byte(tt.args.content), tt.args.key)
			assert.Equal(t, tt.want, got)
		})
	}
}
