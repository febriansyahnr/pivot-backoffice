package platformService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (s *PlatformService) GetSubMerchantUserList(ctx context.Context, request *platform.GetSubMerchantUsersRequest) (*commonModel.PaginationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "/internal/service/v1/platform/GetSubMerchantUserList")
	defer segment.End()

	var (
		responseData []*platform.SubMerchantUserResponse
	)

	err := s.merchantService.ValidateSubMerchantParent(ctx, request.ParentMerchantId, request.MerchantId)
	if err != nil {
		return nil, err
	}

	listRequest := &user.ListUsersByMerchantIDRequest{
		MerchantID: request.MerchantId,
		Name:       request.Keyword,
		RoleID:     request.RoleId,
		SortBy:     request.SortBy,
		SortOrder:  request.SortOrder,
	}
	data, err := s.userService.ListUsersByMerchantID(ctx, listRequest, request.Page, request.PerPage)
	if err != nil {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrGetSubMerchantUserList)
	}

	users, ok := data.Data.([]*user.User)
	if !ok {
		return nil, errPkg.New(response.HttpErrInternal, constant.ErrBuildSubMerchantUserListResponse)
	}

	for _, user := range users {
		subMerchantUser := platform.ToSubMerchantUserResponse(user)
		responseData = append(responseData, subMerchantUser)
	}
	data.Data = responseData

	return data, nil
}
