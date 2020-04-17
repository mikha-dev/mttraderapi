package controller

import (
	"errors"
	"mtdealer"
	"mterr"
	"mtmanapi"
	"mttraderapi/httputil"
	"mttraderapi/model"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListUserTrades godoc
// @Summary List an user trades
// @Description get User trades by login
// @Tags trades
// @Accept  json
// @Produce  json
// @Param login path int true "User ID"
// @Success 200 {object} model.Trade
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router /trades/{login} [get]
func (c *Controller) ListUserTrades(ctx *gin.Context) {

	o, _ := ctx.Get(model.KEY_LOGIN)

	user := o.(model.User)

	mng := getMarketer(ctx)
	if mng == nil {
		return
	}

	trades := mng.GetTrades(func(t *mtdealer.Trade) bool {
		return t.Login == user.Login
	})
	ctx.JSON(http.StatusOK, trades)
}

// AddTrade godoc
// @Summary Add an trade
// @Description add by json trade
// @Tags trades
// @Accept  json
// @Produce  json
// @Param user body model.AddTrade true "Add trade"
// @Success 200 {object} model.Trade
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router /trades/add [post]
func (c *Controller) AddTrade(ctx *gin.Context) {
	o, _ := ctx.Get(model.KEY_LOGIN)
	user := o.(model.User)

	var addTrade model.AddTrade
	if err := ctx.ShouldBindJSON(&addTrade); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := addTrade.Validation(); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}

	dealer := getDealer(ctx)
	market := getMarketer(ctx)
	if dealer == nil || market == nil {
		return
	}

	if !market.IsTradable(user.Login, addTrade.Symbol) {
		httputil.NewError(ctx, http.StatusBadRequest, mterr.InvalidSymbol)
		return
	}

	quote := market.GetQuote(addTrade.Symbol)
	if quote == nil {
		httputil.NewError(ctx, http.StatusBadRequest, mterr.NewInternalError("No market price for %s", addTrade.Symbol))
		return
	}

	price := addTrade.Price

	if addTrade.Command == mtmanapi.OP_BUY {
		price = quote.Ask
	} else if addTrade.Command == mtmanapi.OP_SELL {
		price = quote.Bid
	}

	trade, err := dealer.TradeTransaction(&mtdealer.TradeTrans{
		OrderBy: user.Login,
		Type:    mtmanapi.TT_BR_ORDER_OPEN,
		Cmd:     addTrade.Command,
		Symbol:  addTrade.Symbol,
		Volume:  int(addTrade.Volume * 10),
		Price:   price,
		Sl:      addTrade.Sl,
		Tp:      addTrade.Tp,
	})

	if err != nil {
		httputil.NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, trade)
}

// UpdateTrade godoc
// @Summary Update an trade
// @Description update by json trade
// @Tags trades
// @Accept  json
// @Produce  json
// @Param user body model.UpdateTrade true "Update trade"
// @Success 200 {object} httputil.HTTPError
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router /trades/update [patch]
func (c *Controller) UpdateTrade(ctx *gin.Context) {
	o, _ := ctx.Get(model.KEY_LOGIN)
	user := o.(model.User)

	var updateTrade model.UpdateTrade
	if err := ctx.ShouldBindJSON(&updateTrade); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := updateTrade.Validation(); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}

	dealer := getDealer(ctx)
	market := getMarketer(ctx)
	if dealer == nil || market == nil {
		return
	}

	trade := market.GetTrade(updateTrade.Ticket)
	if trade == nil || trade.Login != user.Login {
		httputil.NewError(ctx, http.StatusBadRequest, mterr.InvalidTrade)
		return
	}

	_, err := dealer.ModifyTrade(&mtdealer.TradeTrans{
		Order: updateTrade.Ticket,
		Price: updateTrade.Price,
		Sl:    updateTrade.Sl,
		Tp:    updateTrade.Tp,
	})

	if err != nil {
		httputil.NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	httputil.NewError(ctx, http.StatusOK, errors.New("updated"))
}

// CloseTrade godoc
// @Summary Close an trade
// @Description close by json trade
// @Tags trades
// @Accept  json
// @Produce  json
// @Param user body model.CloseTrade true "Close trade"
// @Success 200 {object} httputil.HTTPError
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Security ApiKeyAuth
// @Router /trades/close [patch]
func (c *Controller) CloseTrade(ctx *gin.Context) {
	o, _ := ctx.Get(model.KEY_LOGIN)
	user := o.(model.User)

	var closeTrade model.CloseTrade
	if err := ctx.ShouldBindJSON(&closeTrade); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}
	if err := closeTrade.Validation(); err != nil {
		httputil.NewError(ctx, http.StatusBadRequest, err)
		return
	}

	dealer := getDealer(ctx)
	market := getMarketer(ctx)
	if dealer == nil || market == nil {
		return
	}

	trade := market.GetTrade(closeTrade.Ticket)
	if trade == nil || trade.Login != user.Login {
		httputil.NewError(ctx, http.StatusBadRequest, mterr.InvalidTrade)
		return
	}

	quote := market.GetQuote(trade.Symbol)
	if trade == nil {
		httputil.NewError(ctx, http.StatusBadRequest, mterr.InvalidTrade)
		return
	}

	var price float64
	if trade.Cmd == mtmanapi.OP_BUY {
		price = quote.Bid
	} else if trade.Cmd == mtmanapi.OP_SELL {
		price = quote.Ask
	}

	_, err := dealer.TradeTransaction(
		&mtdealer.TradeTrans{
			Order:  trade.Order,
			Type:   mtmanapi.TT_BR_ORDER_CLOSE,
			Volume: int(closeTrade.Volume * 10),
			Price:  price,
		})

	if err != nil {
		httputil.NewError(ctx, http.StatusInternalServerError, err)
		return
	}

	httputil.NewError(ctx, http.StatusOK, errors.New("closed"))
}
