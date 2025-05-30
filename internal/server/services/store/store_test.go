package store

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/LekcRg/metrics/internal/config"
	"github.com/LekcRg/metrics/internal/mocks"
	"github.com/LekcRg/metrics/internal/server/storage"
	"github.com/stretchr/testify/require"
)

func TestSave(t *testing.T) {
	tests := []struct {
		name        string
		dbErr       bool
		wantErr     bool
		changeChmod bool
		chmod       os.FileMode
		db          storage.Database
	}{
		{
			name: "Positive",
			db: storage.Database{
				Gauge: storage.GaugeCollection{
					"gauge1": 123.45,
					"gauge2": 11,
				},
				Counter: storage.CounterCollection{
					"counter1": 123,
					"counter2": 1,
				},
			},
		},
		{
			name:    "Negative db return error",
			dbErr:   true,
			wantErr: true,
			db:      storage.Database{},
		},
		{
			name:        "No access to the file",
			changeChmod: true,
			wantErr:     true,
			chmod:       0o000,
			db:          storage.Database{},
		},
		{
			name:        "Read only access to the file",
			changeChmod: true,
			wantErr:     true,
			chmod:       0o444,
			db:          storage.Database{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "test_save")
			defer func() {
				err = file.Close()
				require.NoError(t, err)
			}()
			require.NoError(t, err)

			st := mocks.NewMockStorage(t)
			var dbErr error
			if tt.dbErr {
				dbErr = errors.New("db err")
			}
			st.EXPECT().GetAll(context.Background()).Return(tt.db, dbErr)

			s := Store{
				cfg: config.ServerConfig{
					FileStoragePath: file.Name(),
				},
				db: st,
			}
			if tt.changeChmod {
				err = os.Chmod(file.Name(), tt.chmod)
				require.NoError(t, err)
			}

			err = s.Save(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRestore(t *testing.T) {
	tests := []struct {
		name        string
		dbErr       bool
		wantErr     bool
		changeChmod bool
		fileContent []byte
		chmod       os.FileMode
		db          storage.Database
	}{
		{
			name: "Positive",
			db: storage.Database{
				Gauge: storage.GaugeCollection{
					"gauge1": 123.45,
					"gauge2": 11,
				},
				Counter: storage.CounterCollection{
					"counter1": 123,
					"counter2": 1,
				},
			},
		},
		{
			name:    "Negative db return error",
			dbErr:   true,
			wantErr: true,
		},
		{
			name:        "No access to the file",
			changeChmod: true,
			wantErr:     true,
			chmod:       0o000,
		},
		{
			name:        "Read only access to the file",
			changeChmod: true,
			wantErr:     false,
			chmod:       0o444,
		},
		{
			name:        "Invalid JSON",
			fileContent: []byte("invalid}}"),
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.CreateTemp("", "test_save")
			require.NoError(t, err)
			defer file.Close()

			var fileContent []byte
			if len(tt.fileContent) > 0 {
				fileContent = tt.fileContent
			} else {
				fileContent, err = json.Marshal(tt.db)
				require.NoError(t, err)
			}

			_, err = file.Write(fileContent)
			require.NoError(t, err)

			st := mocks.NewMockStorage(t)
			if !tt.wantErr || tt.dbErr {
				var dbErr error
				if tt.dbErr {
					dbErr = errors.New("db err")
				}
				st.EXPECT().UpdateMany(context.Background(), tt.db).Return(dbErr)
			}

			s := Store{
				cfg: config.ServerConfig{
					FileStoragePath: file.Name(),
				},
				db: st,
			}
			if tt.changeChmod {
				err = os.Chmod(file.Name(), tt.chmod)
				require.NoError(t, err)
			}

			err = s.Restore(context.Background())
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
