package menuService

import (
	"context"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	productModel "github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *MenuService) GetByRole(ctx context.Context, roleID string, isMenuFormatting bool) ([]*menuModel.MenuResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/menu/GetByRole")
	defer segment.End()

	menus, err := s.repo.GetAll(ctx, &menuModel.GetAllFilterRequest{RoleID: roleID})
	if err != nil {
		return nil, err
	}

	// Exclude disallowed products
	menus = s.excludeDisallowedProducts(ctx, menus)

	// If user has required permission, add Home menu
	if hasRequiredPermissionForHome(menus) {
		homeMenu, err := s.repo.FindBySlugWithPermissions(ctx, "home")
		if err != nil {
			s.logger.Warn(ctx, "failed to fetch Home menu", logger.Error(err))
		} else if homeMenu != nil {
			menus = append([]*menuModel.MenuResponse{homeMenu}, menus...)
		}
	}

	if !isMenuFormatting {
		return menus, nil
	}

	list := groupByParent(menus)
	return createMenuTree(list, list[""]), nil
}

// groupByParent groups menu items by their parent ID
func groupByParent(items []*menuModel.MenuResponse) map[string][]*menuModel.MenuResponse {
	list := make(map[string][]*menuModel.MenuResponse)
	for _, item := range items {
		parent := ""
		if item.ParentID != nil {
			parent = *item.ParentID
		}

		list[parent] = append(list[parent], item)
	}
	return list
}

// createMenuTree builds a tree from a flat list of menu items
func createMenuTree(list map[string][]*menuModel.MenuResponse, parent []*menuModel.MenuResponse) []*menuModel.MenuResponse {
	var tree []*menuModel.MenuResponse
	for _, item := range parent {
		if children, exists := list[item.UUID]; exists {
			item.Children = createMenuTree(list, children)
		}
		tree = append(tree, item)
	}
	return tree
}

// hasRequiredPermissionForHome checks if any menu has a permission that grants access to Home menu
func hasRequiredPermissionForHome(menus []*menuModel.MenuResponse) bool {
	for _, menu := range menus {
		for _, perm := range menu.Permissions {
			for _, prefix := range constant.HomeMenuRequiredPermissionPrefixes {
				if strings.HasPrefix(perm.Slug, prefix) {
					return true
				}
			}
		}
	}
	return false
}

func (s *MenuService) excludeDisallowedProducts(ctx context.Context, menus []*menuModel.MenuResponse) []*menuModel.MenuResponse {
	ctx, span := otelTracer.Start(ctx, "internal/service/v1/menu/excludeDisallowedProducts")
	defer span.End()

	// Get Merchant From Context If Any
	merchantID, _ := ctx.Value(constant.CtxMerchantIDKey).(string)

	filtered := menus[:0]
	for _, menu := range menus {
		if menu.AllowedProducts == nil || merchantID == "" {
			filtered = append(filtered, menu)
			continue
		}

		if !s.hasAllowedProducts(ctx, merchantID, menu) {
			continue
		}

		filtered = append(filtered, menu)
	}

	return filtered
}

func (s *MenuService) hasAllowedProducts(
	ctx context.Context,
	merchantID string,
	menu *menuModel.MenuResponse,
) bool {
	for _, allowedProduct := range *menu.AllowedProducts {
		if err := s.productService.ValidateMerchantProductAvailability(
			ctx,
			&productModel.ValidateMerchantProductAvailability{
				MerchantID:  merchantID,
				ProductName: allowedProduct,
			},
		); err != nil {
			return false
		}
	}

	return true
}
