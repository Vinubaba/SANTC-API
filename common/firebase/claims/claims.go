package claims

import (
	"context"

	"github.com/Vinubaba/SANTC-API/common/roles"
	"github.com/Vinubaba/SANTC-API/common/store"
)

func IsAdmin(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims != nil && claims[roles.ROLE_ADMIN] != nil {
		return claims[roles.ROLE_ADMIN].(bool)
	}
	return false
}

func IsOfficeManager(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims != nil && claims[roles.ROLE_OFFICE_MANAGER] != nil {
		return claims[roles.ROLE_OFFICE_MANAGER].(bool)
	}
	return false
}

func IsTeacher(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims != nil && claims[roles.ROLE_TEACHER] != nil {
		return claims[roles.ROLE_TEACHER].(bool)
	}
	return false
}

func IsAdult(ctx context.Context) bool {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims != nil && claims[roles.ROLE_ADULT] != nil {
		return claims[roles.ROLE_ADULT].(bool)
	}
	return false
}

func GetDaycareId(ctx context.Context) string {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims != nil && claims["daycareId"] != nil {
		return claims["daycareId"].(string)
	}
	return ""
}

func GetUserId(ctx context.Context) string {
	claims := ctx.Value("claims").(map[string]interface{})
	if claims != nil && claims["userId"] != nil {
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
