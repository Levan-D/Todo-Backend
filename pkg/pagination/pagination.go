package pagination

import (
	"fmt"
	"github.com/Levan-D/Todo-Backend/pkg/domain"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type Response struct {
	Data interface{} `json:"data"`
	Meta *Meta       `json:"_meta"`
}

type Meta struct {
	Limit   int      `json:"limit"`
	Current int      `json:"current"`
	Sorter  string   `json:"sorter"`
	Total   int64    `json:"total"`
	Search  []Search `json:"search"`
}

type Search struct {
	Column string `json:"column"`
	Action string `json:"action"`
	Query  string `json:"query"`
}

type Input struct {
	Limit     int
	Current   int
	Sorter    string
	Relations []string
	Query     string
}

type Query struct {
	Limit     string `json:"limit"`
	Current   string `json:"current"`
	Sorter    string `json:"sorter"`
	Relations string `json:"relations"`
}

func Parse(c *fiber.Ctx) Input {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	current, _ := strconv.Atoi(c.Query("current", "1"))
	sorter := c.Query("sorter", "id asc")
	relations := strings.Split(c.Query("relations"), ",")
	query := string(c.Request().URI().QueryString())

	return Input{
		Limit:     limit,
		Current:   current,
		Sorter:    sorter,
		Relations: relations,
		Query:     query,
	}
}

type DataType string

const (
	DataTypeNone        DataType = "NONE"
	DataTypePayment     DataType = "PAYMENT"
	DataTypeTransaction DataType = "TRANSACTION"
)

func FindAll(input Input, db *gorm.DB, dest interface{}, where map[string]string, dataType DataType) (Response, error) {
	var searchs []Search
	// default limit, page & sort parameter
	limit := input.Limit
	current := input.Current
	sorter := input.Sorter
	relations := input.Relations

	// decode query parameters
	decodedQuery, _ := url.QueryUnescape(input.Query)
	searchQuery := strings.Split(decodedQuery, "&")

	for _, item := range searchQuery {
		query := strings.Split(item, "=")
		// check if query parameter key contains dot
		if strings.Contains(query[0], "@") {
			// split query parameter key by dot
			searchKeys := strings.Split(query[0], "@")
			// create search object
			search := Search{Column: searchKeys[0], Action: searchKeys[1], Query: query[1]}
			// add search object to searchs array
			searchs = append(searchs, search)
		}
	}

	// setup current page and offset
	current -= 1
	offset := current * limit
	// get data with limit, offset & order
	find := db.Order(sorter)
	findCount := db.Model(dest)
	// Relations
	if len(relations) > 0 && relations[0] != "" {
		for _, item := range relations {
			find.Preload(item)
		}
	}
	// add limit and offset for query
	find = find.Limit(limit).Offset(offset)
	// add search for query
	if searchs != nil {
		for _, value := range searchs {
			column := value.Column
			action := value.Action
			query := value.Query

			switch action {
			case "equals":
				whereQuery := fmt.Sprintf("%s = ?", column)
				find = find.Where(whereQuery, query)
				findCount = findCount.Where(whereQuery, query)
				break
			case "contains":
				if strings.Contains(column, ".") && reflect.TypeOf(dest) == reflect.TypeOf((&[]domain.TenantInvoice{})) {
					sp := strings.Split(column, ".")
					whereQuery := fmt.Sprintf("EXISTS (SELECT FROM %s WHERE id = tenant_invoices.tenant_id AND LOWER(%s) LIKE LOWER(?))", sp[0], column)
					find = find.Where(whereQuery, "%"+query+"%")
					findCount = findCount.Where(whereQuery, "%"+query+"%")
				} else if strings.Contains(column, ".") && reflect.TypeOf(dest) == reflect.TypeOf((&[]domain.Payment{})) {
					sp := strings.Split(column, ".")
					whereQuery := fmt.Sprintf("EXISTS (SELECT FROM %s WHERE LOWER(%s) LIKE LOWER(?))", sp[0], column)
					find = find.Where(whereQuery, "%"+query+"%")
					findCount = findCount.Where(whereQuery, "%"+query+"%")
				} else {
					whereQuery := fmt.Sprintf("LOWER(%s) LIKE LOWER(?)", column)
					find = find.Where(whereQuery, "%"+query+"%")
					findCount = findCount.Where(whereQuery, "%"+query+"%")
				}

				/*
					 else if strings.Contains(column, ".") && reflect.TypeOf(dest) == reflect.TypeOf((&[]domain.UserQuestionnaireAnswer{})) {
							sp := strings.Split(column, ".")
							whereQuery := fmt.Sprintf("EXISTS (SELECT FROM %s WHERE id = question.question_id AND LOWER(%s) LIKE LOWER(?))", sp[0], column)
							find = find.Where(whereQuery, "%"+query+"%")
							findCount = findCount.Where(whereQuery, "%"+query+"%")
						}
				*/
				break
			case "in":
				whereQuery := fmt.Sprintf("%s IN (?)", column)
				queryArray := strings.Split(query, ",")
				find = find.Where(whereQuery, queryArray)
				findCount = findCount.Where(whereQuery, queryArray)
				break
			case "between":
				whereQuery := fmt.Sprintf("%s BETWEEN ? AND ?", column)
				query := strings.Split(query, ",")
				find = find.Where(whereQuery, query[0], query[1])
				findCount = findCount.Where(whereQuery, query[0], query[1])
				break
			case "notnull":
				whereQuery := fmt.Sprintf("%s IS NOT NULL", column)
				find = find.Where(whereQuery)
				findCount = findCount.Where(whereQuery)
				break
			case "bool":
				whereQuery := fmt.Sprintf("%s = ?", column)
				status, _ := strconv.ParseBool(query)
				find = find.Where(whereQuery, status)
				findCount = findCount.Where(whereQuery, status)
				break
			}
		}
	}
	// where variable init
	for key, item := range where {
		find.Where(key+" = ?", item)
		findCount = findCount.Where(key+" = ?", item)
	}

	if reflect.TypeOf(dest).String() == "*[]domain.Payment" {
		if dataType == DataTypePayment {
			find.Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment)
			find.Where("status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned)
			findCount.Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment)
			findCount.Where("status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned)
		} else if dataType == DataTypeTransaction {
			find.Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment)
			find.Where("status = ? OR status = ? OR status = ? OR status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned, domain.PaymentStatusCancelled, domain.PaymentStatusDeclined, domain.PaymentStatusPending)
			findCount.Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment)
			findCount.Where("status = ? OR status = ? OR status = ? OR status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned, domain.PaymentStatusCancelled, domain.PaymentStatusDeclined, domain.PaymentStatusPending)
		} else {
			find.Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment)
			find.Where("status = ? OR status = ? OR status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned, domain.PaymentStatusCancelled, domain.PaymentStatusDeclined)
			findCount.Where("type = ? OR type = ? OR type = ?", domain.PaymentTypeTenantPayment, domain.PaymentTypeTenantReturn, domain.PaymentTypeParkingPayment)
			findCount.Where("status = ? OR status = ? OR status = ? OR status = ?", domain.PaymentStatusPaid, domain.PaymentStatusReturned, domain.PaymentStatusCancelled, domain.PaymentStatusDeclined)
		}
	}

	// find a query
	find = find.Find(dest)

	// has error find data
	errFind := find.Error
	if errFind != nil {
		return Response{}, errFind
	}
	// count all data
	var total int64
	errCount := findCount.Count(&total).Error
	if errCount != nil {
		return Response{}, errCount
	}

	// return collected data
	return Response{
		Data: dest,
		Meta: &Meta{
			Limit:   limit,
			Current: current + 1,
			Sorter:  sorter,
			Search:  searchs,
			Total:   total,
		},
	}, nil
}
