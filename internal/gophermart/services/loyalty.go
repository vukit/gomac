package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/vukit/gomac/internal/gophermart/logger"
	"github.com/vukit/gomac/internal/gophermart/models"
	"github.com/vukit/gomac/internal/gophermart/repositories"
)

type LoyaltyService struct {
	Address string
	Repo    repositories.Repo
	Logger  *logger.Logger
}

func (r *LoyaltyService) EarnPoints(ctx context.Context, task models.Task) {
	var lsData struct {
		Order   string
		Status  string
		Accrual float64
	}

	client := &http.Client{}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.Address+"/api/orders/"+task.OrderNumber, &bytes.Buffer{})
	if err != nil {
		r.Logger.Warning(err.Error())

		return
	}

	for {
		resp, err := client.Do(req)
		if err != nil {
			r.Logger.Warning(err.Error())
			time.Sleep(2 * time.Second)

			continue
		}

		statusCode := resp.StatusCode
		if statusCode != http.StatusOK {
			err = fmt.Errorf("loyalty service status code %d for order %s", statusCode, task.OrderNumber)
			r.Logger.Warning(err.Error())
			time.Sleep(2 * time.Second)

			continue
		}

		decoder := json.NewDecoder(resp.Body)

		err = decoder.Decode(&lsData)
		if err != nil {
			resp.Body.Close()
			r.Logger.Warning(err.Error())

			continue
		}

		resp.Body.Close()

		if task.Accrual != lsData.Accrual || task.Status != lsData.Status {
			task.Accrual = lsData.Accrual
			task.Status = lsData.Status

			err := r.Repo.SaveTask(ctx, task)
			if err != nil {
				r.Logger.Warning(err.Error())
			}
		}

		if task.Status == "PROCESSED" || task.Status == "INVALID" {
			return
		}
	}
}
