package abstract

/*
 * @Desc: 校验
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhtranslations "github.com/go-playground/validator/v10/translations/zh"
	"reflect"
)

var ValidateServiceApi = newValidateService()

type validateService struct {
	validate *validator.Validate
	trans    ut.Translator
}

type errorResult struct {
	JsonName string `json:"json"`
	ErrMsg   string `json:"errMsg"`
}

func newValidateService() *validateService {
	zh := zh.New()
	uni := ut.New(zh, zh)
	validate := validator.New()
	trans, _ := uni.GetTranslator("zh")
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := field.Tag.Get("label")
		if name == "" {
			name = field.Name
		}
		return name
	})
	zhtranslations.RegisterDefaultTranslations(validate, trans)
	return &validateService{validate: validate, trans: trans}
}

func (vd *validateService) Validate(model any) []errorResult {
	var errors []errorResult
	if model == nil {
		errors = append(errors, errorResult{ErrMsg: "验证对象为空."})
		return errors
	}
	err := vd.validate.Struct(model)
	if err != nil {
		tv := reflect.TypeOf(model)
		if tv.Kind() == reflect.Ptr {
			tv = tv.Elem()
		}
		jsonFieldMap := map[string]string{}
		for i := 0; i < tv.NumField(); i++ {
			jsonFieldMap[tv.Field(i).Name] = tv.Field(i).Tag.Get("json")
		}
		errs := err.(validator.ValidationErrors)
		for _, fieldError := range errs {
			errors = append(errors, errorResult{ErrMsg: fieldError.Translate(vd.trans),
				JsonName: jsonFieldMap[fieldError.StructField()],
			})
		}
	}
	return errors
}
