package common

type Paging struct {
	PageIndex  int    `json:"pageindex" form:"pageindex"`
	PageSize   int    `json:"pagesize" form:"pagesize"`
	TotalCount int    `json:"totalcount" form:"totalcount"`
	FakeCursor string `json:"fakecursor" form:"fakecursor"`
	NextCursor string `json:"nextcursor" form:"nextcursor"`
}

func (p *Paging) Process() {
	if p.PageIndex < 1 {
		p.PageIndex = 1
	}

	if p.PageSize < 1 {
		p.PageSize = 10
	}

	if p.PageSize >= 100 {
		p.PageSize = 100
	}
}
