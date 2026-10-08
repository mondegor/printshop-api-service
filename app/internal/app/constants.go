package app

const (
	// PageSizeDefault - размер страницы списков по умолчанию (согласован с контрактами
	// Api.Request.Query.ListCursor.Limit и Api.Request.Query.ListPager.PageSize).
	PageSizeDefault = 50

	// PageSizeMax - предельный размер страницы списков (согласован с теми же контрактами).
	PageSizeMax = 1000
)
