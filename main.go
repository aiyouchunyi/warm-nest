package main

import (
	"warm-nest/pkg/app"
	widgets2 "warm-nest/pkg/app/widgets"
	"warm-nest/pkg/tool/auth"
	"warm-nest/pkg/tool/auth/jwt"
	"warm-nest/pkg/tool/database"
	"warm-nest/pkg/tool/database/mysqls/widgets"
	"warm-nest/pkg/tool/tasks"

	"warm-nest/internal"
	"warm-nest/internal/provider"
	"warm-nest/internal/service"
)

func main() {
	app.New("warm-nest").
		DB(database.MysqlDriver).
		Web(internal.RegisterController()).
		Do(internal.EnableLoader).
		Do(widgets.EnableModel, internal.RegisterModel()).
		Do(auth.EnableAuth, jwt.GetJWTService(), service.GetAuthService(), provider.GetUserProvider()).
		DoCtx(tasks.EnableTask, internal.RegisterTask()).
		AsyncDo(widgets2.EnableMigrate, internal.RegisterMigrate()).
		Run()
}
