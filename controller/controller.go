package controller

import (
	"errors"
	"mtdealer"
	"mttraderapi/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Controller example
type Controller struct {
}

func getMarketer(ctx *gin.Context) *mtdealer.MarketManager {
	val, ok := ctx.Get(model.KEY_MANAGER)

	if !ok {
		_ = ctx.AbortWithError(http.StatusInternalServerError, errors.New("Marketer not found in context"))
		return nil
	}

	if mng, ok := val.(*mtdealer.MarketManager); ok {
		return mng
	}

	_ = ctx.AbortWithError(http.StatusInternalServerError, errors.New("Marketer not found in context"))
	return nil
}

func getDealer(ctx *gin.Context) *mtdealer.DealerManager {
	val, ok := ctx.Get(model.KEY_DEALER)

	if !ok {
		_ = ctx.AbortWithError(http.StatusInternalServerError, errors.New("Dealer not found in context"))
		return nil
	}

	if dealer, ok := val.(*mtdealer.DealerManager); ok {
		return dealer
	}

	_ = ctx.AbortWithError(http.StatusInternalServerError, errors.New("Dealer not found in context"))
	return nil
}

// NewController example
func NewController() *Controller {
	return &Controller{}
}
