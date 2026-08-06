package liveFeature

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/liveFeature"
	rmqMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {
	testCases := []struct {
		name     string
		mockFunc func(mockRepo *mocks.ILiveFeatureRepository)
		want     []liveFeature.LiveFeature
		wantErr  bool
	}{
		{
			name: "Success",
			mockFunc: func(mockRepo *mocks.ILiveFeatureRepository) {
				mockRepo.On("GetAll", mock.Anything).Return([]liveFeature.LiveFeature{
					{UUID: "uuid-uuid-uuid", Name: "Feature 1"},
					{UUID: "uuid-uuid-uuid-uuid", Name: "Feature 2"},
				}, nil)
			},
			want: []liveFeature.LiveFeature{
				{UUID: "uuid-uuid-uuid", Name: "Feature 1"},
				{UUID: "uuid-uuid-uuid-uuid", Name: "Feature 2"},
			},
			wantErr: false,
		},
		{
			name: "Repository error",
			mockFunc: func(mockRepo *mocks.ILiveFeatureRepository) {
				mockRepo.On("GetAll", mock.Anything).Return(nil, errors.New("repository error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Empty list",
			mockFunc: func(mockRepo *mocks.ILiveFeatureRepository) {
				mockRepo.On("GetAll", mock.Anything).Return([]liveFeature.LiveFeature{}, nil)
			},
			want:    []liveFeature.LiveFeature{},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := mocks.NewILiveFeatureRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			rmq := rmqMock.NewRabbitMQExt(t)

			tc.mockFunc(mockRepo)

			svc := New(mockLogger, mockRepo, rmq)

			got, err := svc.GetList(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tc.want, got)
			mockRepo.AssertExpectations(t)
		})
	}
}
