package period

import "github.com/mysecodgit/go_accounting/src/building"

type PeriodResponse struct {
	ID         int               `json:"id"`
	PeriodName string            `json:"period_name"`
	Start      string            `json:"start"`
	End        string            `json:"end"`
	Building   building.Building `json:"building"`
	IsClosed   int               `json:"is_closed"`
	CreatedAt  string            `json:"created_at"`
	UpdatedAt  string            `json:"updated_at"`
}

func (p *Period) ToPeriodResponse(b building.Building) PeriodResponse {
	return PeriodResponse{
		ID:         p.ID,
		PeriodName: p.PeriodName,
		Start:      p.Start,
		End:        p.End,
		Building:   b,
		IsClosed:   p.IsClosed,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}
