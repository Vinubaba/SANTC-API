package claims

import (
	"context"

	"github.com/Vinubaba/SANTC-API/api/shared"
	"github.com/Vinubaba/SANTC-API/api/store"
)

func IsAdmin(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	return claims[shared.ROLE_ADMIN].(bool)
}

func IsOfficeManager(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	return claims[shared.ROLE_OFFICE_MANAGER].(bool)
}

func IsTeacher(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	return claims[shared.ROLE_TEACHER].(bool)
}

func IsAdult(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	return claims[shared.ROLE_ADULT].(bool)
}

func GetDaycareId(ctx context.Context) string {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims["daycareId"] != nil {
		return claims["daycareId"].(string)
	}
	return ""
}

func GetUserId(ctx context.Context) string {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims["userId"] != nil {
		return claims["userId"].(string)
	}
	return ""
}

func GetDefaultSearchOptions(ctx context.Context) store.SearchOptions {
	searchOptions := store.SearchOptions{}
	daycareId := GetDaycareId(ctx)
	userId := GetUserId(ctx)

	if !IsAdmin(ctx) {
		searchOptions.DaycareId = daycareId
	}
	if IsAdult(ctx) {
		searchOptions.ResponsibleId = userId
	}
	if IsTeacher(ctx) {
		searchOptions.TeacherId = userId
	}

	return searchOptions
}
