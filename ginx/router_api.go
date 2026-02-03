package ginx

import (
	"mime/multipart"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-xuan/utilx/excelx"
	"github.com/go-xuan/utilx/idx"
	"gorm.io/gorm"
)

// BindCrudRouter 新增crud路由
func BindCrudRouter[T any](router *gin.RouterGroup, db *gorm.DB) {
	api := &Model[T]{db: db}
	router.GET("list", api.List)        // 列表
	router.GET("detail", api.Detail)    // 明细
	router.POST("create", api.Create)   // 新增
	router.PUT("update", api.Update)    // 修改
	router.DELETE("delete", api.Delete) // 删除
}

// BindExcelRouter 新增 Excel 相关路由
func BindExcelRouter[T any](group *gin.RouterGroup, db *gorm.DB) {
	api := &Model[T]{db: db}
	group.POST("import", api.Import) // 导入
	group.POST("export", api.Export) // 导出
}

// Model 通用模型
type Model[T any] struct {
	db *gorm.DB
}

// GetDB 获取数据库连接
func (m *Model[T]) GetDB() *gorm.DB {
	return m.db
}

// List 列表
func (m *Model[T]) List(ctx *gin.Context) {
	var result []*T
	if err := m.GetDB().Find(&result).Error; err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, result)
}

// Create 新增
func (m *Model[T]) Create(ctx *gin.Context) {
	var err error
	var create T
	if err = ctx.ShouldBindJSON(&create); err != nil {
		ParamError(ctx, err)
		return
	}
	if err = m.GetDB().Create(&create).Error; err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, nil)
}

// Update 修改
func (m *Model[T]) Update(ctx *gin.Context) {
	var update T
	if err := ctx.ShouldBindJSON(&update); err != nil {
		ParamError(ctx, err)
		return
	}
	if err := m.GetDB().Updates(&update).Error; err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, nil)
}

// Delete 删除
func (m *Model[T]) Delete(ctx *gin.Context) {
	var req struct {
		Id string `form:"id" json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ParamError(ctx, err)
		return
	}
	var t T
	if err := m.GetDB().Where("id = ? ", req.Id).Delete(&t).Error; err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, nil)
}

// Detail 明细
func (m *Model[T]) Detail(ctx *gin.Context) {
	var req struct {
		Id string `form:"id" json:"id" binding:"required"`
	}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ParamError(ctx, err)
		return
	}
	var result T
	if err := m.GetDB().Where("id = ? ", req.Id).Find(&result).Error; err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, result)
}

// Import 导入
func (m *Model[T]) Import(ctx *gin.Context) {
	var file struct {
		File *multipart.FileHeader `form:"file"`
	}
	if err := ctx.ShouldBind(&file); err != nil {
		ParamError(ctx, err)
		return
	}
	path := filepath.Join("import", file.File.Filename)
	if err := ctx.SaveUploadedFile(file.File, path); err != nil {
		ParamError(ctx, err)
		return
	}
	var t T
	if data, err := excelx.ReadAny(path, "", t); err != nil {
		CustomResponse(ctx, NewResponse(ExportFailedCode, err.Error()))
		return
	} else if err = m.GetDB().Model(t).Create(&data).Error; err != nil {
		Error(ctx, err)
		return
	}
	Success(ctx, nil)
}

// Export 导出
func (m *Model[T]) Export(ctx *gin.Context) {
	var result []*T
	if err := m.GetDB().Find(&result).Error; err != nil {
		CustomResponse(ctx, NewResponse(ExportFailedCode, err.Error()))
		return
	}

	filePath := filepath.Join("export", idx.Timestamp()+".xlsx")
	if len(result) > 0 {
		if err := excelx.WriteAny(filePath, result); err != nil {
			CustomResponse(ctx, NewResponse(ExportFailedCode, err.Error()))
			return
		}
	}
	ctx.File(filePath)
}
