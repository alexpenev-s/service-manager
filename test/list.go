package test

import (
	"encoding/json"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/extensions/table"

	"net/http"

	. "github.com/onsi/ginkgo"

	"github.com/Peripli/service-manager/test/common"
)

type listOpEntry struct {
	expectedResourcesBeforeOp []common.Object

	queryTemplate              string
	queryArgs                  common.Object
	expectedResourcesAfterOp   []common.Object
	unexpectedResourcesAfterOp []common.Object
	expectedStatusCode         int
}

func DescribeListTestsFor(ctx *common.TestContext, t TestCase, r []common.Object, rWithMandatoryFields common.Object) bool {
	entries := []TableEntry{
		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp: []common.Object{r[0]},
				queryTemplate:             "%s = %v",
				queryArgs:                 r[0],
				expectedResourcesAfterOp:  []common.Object{r[0]},
				expectedStatusCode:        http.StatusOK,
			},
		),
		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp:  []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:              "%s != %v",
				queryArgs:                  r[0],
				unexpectedResourcesAfterOp: []common.Object{r[0]},
				expectedStatusCode:         http.StatusOK,
			},
		),

		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp: []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:             "%[1]s in [%[2]v||%[2]v||%[2]v]",
				queryArgs:                 r[0],
				expectedResourcesAfterOp:  []common.Object{r[0]},
				expectedStatusCode:        http.StatusOK,
			},
		),

		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp: []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:             "%s in [%v]",
				queryArgs:                 r[0],
				expectedResourcesAfterOp:  []common.Object{r[0]},
				expectedStatusCode:        http.StatusOK,
			},
		),
		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp:  []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:              "%[1]s notin [%[2]v||%[2]v||%[2]v]",
				queryArgs:                  r[0],
				unexpectedResourcesAfterOp: []common.Object{r[0]},
				expectedStatusCode:         http.StatusOK,
			},
		),
		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp:  []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:              "%s notin [%v]",
				queryArgs:                  r[0],
				unexpectedResourcesAfterOp: []common.Object{r[0]},
				expectedStatusCode:         http.StatusOK,
			},
		),
		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp:  []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:              "%s gt %v",
				queryArgs:                  common.RmNonNumericArgs(r[0]),
				unexpectedResourcesAfterOp: []common.Object{r[0]},
				expectedStatusCode:         http.StatusOK,
			},
		),
		Entry("returns 200",
			listOpEntry{
				expectedResourcesBeforeOp:  []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:              "%s lt %v",
				queryArgs:                  common.RmNonNumericArgs(r[0]),
				unexpectedResourcesAfterOp: []common.Object{r[0]},
				expectedStatusCode:         http.StatusOK,
			},
		),
		Entry("returns 200 for field queries",
			listOpEntry{
				expectedResourcesBeforeOp: []common.Object{r[0], rWithMandatoryFields},
				queryTemplate:             "%s eqornil %v",
				queryArgs:                 common.RmNotNullableFieldAndLabels(r[0], rWithMandatoryFields),
				expectedResourcesAfterOp:  []common.Object{r[0], rWithMandatoryFields},
				expectedStatusCode:        http.StatusOK,
			},
		),
		Entry("returns 400 for label queries with operator eqornil",
			listOpEntry{
				queryTemplate: "%s eqornil %v",
				queryArgs: common.Object{
					"labels": map[string]interface{}{
						"labelKey1": []interface{}{
							"str",
						},
					}},
				expectedStatusCode: http.StatusBadRequest,
			},
		),
		Entry("returns 200 for JSON fields with stripped new lines",
			listOpEntry{
				expectedResourcesBeforeOp: []common.Object{r[0]},
				queryTemplate:             "%s = %v",
				queryArgs:                 common.RmNonJSONArgs(r[0]),
				expectedResourcesAfterOp:  []common.Object{r[0]},
				expectedStatusCode:        http.StatusOK,
			},
		),

		Entry("returns 400 when query operator is invalid",
			listOpEntry{
				queryTemplate:      "%s @@ %v",
				queryArgs:          r[0],
				expectedStatusCode: http.StatusBadRequest,
			},
		),
		Entry("returns 400 when query is duplicated",
			listOpEntry{
				queryTemplate:      "%[1]s = %[2]v|%[1]s = %[2]v",
				queryArgs:          r[0],
				expectedStatusCode: http.StatusBadRequest,
			},
		),
		Entry("returns 400 when operator is not properly separated with right space from operands",
			listOpEntry{
				queryTemplate:      "%s =%v",
				queryArgs:          r[0],
				expectedStatusCode: http.StatusBadRequest,
			},
		),
		Entry("returns 400 when operator is not properly separated with left space from operands",
			listOpEntry{
				queryTemplate:      "%s= %v",
				queryArgs:          r[0],
				expectedStatusCode: http.StatusBadRequest,
			},
		),

		Entry("returns 400 when field query left operands are unknown",
			listOpEntry{
				queryTemplate:      "%[1]s in [%[2]v||%[2]v]",
				queryArgs:          common.Object{"unknownkey": "unknownvalue"},
				expectedStatusCode: http.StatusBadRequest,
			},
		),
		Entry("returns 200 when label query left operands are unknown",
			listOpEntry{
				expectedResourcesBeforeOp: []common.Object{r[0], r[1], r[2], r[3]},
				queryTemplate:             "%[1]s in [%[2]v||%[2]v]",
				queryArgs: common.Object{
					"labels": map[string]interface{}{
						"unknown": []interface{}{
							"unknown",
						},
					}},
				unexpectedResourcesAfterOp: []common.Object{r[0], r[1], r[2], r[3]},
				expectedStatusCode:         http.StatusOK,
			},
		),
		Entry("returns 400 when single value operator is used with multiple right value arguments",
			listOpEntry{
				queryTemplate:      "%[1]s != [%[2]v||%[2]v||%[2]v]",
				queryArgs:          r[0],
				expectedStatusCode: http.StatusBadRequest,
			},
		),

		Entry("returns 400 when numeric operator is used with non-numeric operands",
			listOpEntry{
				queryTemplate:      "%s < %v",
				queryArgs:          common.RmNumericArgs(r[0]),
				expectedStatusCode: http.StatusBadRequest,
			},
		),
	}

	verifyListOp := func(listOpEntry listOpEntry, query string) {
		var expectedAfterOpIDs []string
		var unexpectedAfterOpIDs []string
		expectedAfterOpIDs = common.ExtractResourceIDs(listOpEntry.expectedResourcesAfterOp)
		unexpectedAfterOpIDs = common.ExtractResourceIDs(listOpEntry.unexpectedResourcesAfterOp)

		By(fmt.Sprintf("[TEST]: Verifying expected %s before operation after present", t.API))
		beforeOpArray := ctx.SMWithOAuth.GET("/v1/" + t.API).
			Expect().
			Status(http.StatusOK).JSON().Object().Value(t.API).Array()

		for _, v := range beforeOpArray.Iter() {
			obj := v.Object().Raw()
			delete(obj, "created_at")
			delete(obj, "updated_at")
		}

		for _, entity := range listOpEntry.expectedResourcesBeforeOp {
			delete(entity, "created_at")
			delete(entity, "updated_at")
			beforeOpArray.Contains(entity)
		}

		By("[TEST]: ======= Expectations Summary =======")

		By(fmt.Sprintf("[TEST]: Listing %s with %s", t.API, query))
		By(fmt.Sprintf("[TEST]: Currently present resources: %v", r))
		By(fmt.Sprintf("[TEST]: Expected %s ids after operations: %s", t.API, expectedAfterOpIDs))
		By(fmt.Sprintf("[TEST]: Unexpected %s ids after operations: %s", t.API, unexpectedAfterOpIDs))
		By(fmt.Sprintf("[TEST]: Expected status code %d", listOpEntry.expectedStatusCode))

		By("[TEST]: ====================================")

		req := ctx.SMWithOAuth.GET("/v1/" + t.API)
		if query != "" {
			req = req.WithQueryString(query)
		}

		By(fmt.Sprintf("[TEST]: Verifying expected status code %d is returned from list operation", listOpEntry.expectedStatusCode))
		resp := req.
			Expect().
			Status(listOpEntry.expectedStatusCode)

		if listOpEntry.expectedStatusCode != http.StatusOK {
			By(fmt.Sprintf("[TEST]: Verifying error and description fields are returned after list operation"))

			resp.JSON().Object().Keys().Contains("error", "description")
		} else {
			array := resp.JSON().Object().Value(t.API).Array()
			for _, v := range array.Iter() {
				obj := v.Object().Raw()
				delete(obj, "created_at")
				delete(obj, "updated_at")
			}

			if listOpEntry.expectedResourcesAfterOp != nil {
				By(fmt.Sprintf("[TEST]: Verifying expected %s are returned after list operation", t.API))
				for _, entity := range listOpEntry.expectedResourcesAfterOp {
					delete(entity, "created_at")
					delete(entity, "updated_at")
					array.Contains(entity)
				}
			}

			if listOpEntry.unexpectedResourcesAfterOp != nil {
				By(fmt.Sprintf("[TEST]: Verifying unexpected %s are NOT returned after list operation", t.API))

				for _, entity := range listOpEntry.unexpectedResourcesAfterOp {
					delete(entity, "created_at")
					delete(entity, "updated_at")
					array.NotContains(entity)
				}
			}
		}
	}

	return Describe("List", func() {
		Context("with basic auth", func() {
			It("returns 200", func() {
				ctx.SMWithBasic.GET("/v1/" + t.API).
					Expect().
					Status(http.StatusOK)
			})
		})

		Context("with bearer auth", func() {
			Context("with no query", func() {
				It("returns all the resources", func() {
					verifyListOp(listOpEntry{
						expectedResourcesBeforeOp: []common.Object{r[0], r[1]},
						expectedResourcesAfterOp:  []common.Object{r[0], r[1]},
						expectedStatusCode:        http.StatusOK,
					}, "")
				})
			})

			Context("with empty field query", func() {
				It("returns 200", func() {
					verifyListOp(listOpEntry{
						expectedResourcesBeforeOp: []common.Object{r[0], r[1]},
						expectedResourcesAfterOp:  []common.Object{r[0], r[1]},
						expectedStatusCode:        http.StatusOK,
					}, "fieldQuery=")
				})
			})

			Context("with empty label query", func() {
				It("returns 200", func() {
					verifyListOp(listOpEntry{
						expectedResourcesBeforeOp: []common.Object{r[0], r[1]},
						expectedResourcesAfterOp:  []common.Object{r[0], r[1]},
						expectedStatusCode:        http.StatusOK,
					}, "labelQuery=")
				})
			})

			Context("with empty label query and field query", func() {
				It("returns 200", func() {
					verifyListOp(listOpEntry{
						expectedResourcesBeforeOp: []common.Object{r[0], r[1]},
						expectedResourcesAfterOp:  []common.Object{r[0], r[1]},
						expectedStatusCode:        http.StatusOK,
					}, "labelQuery=&fieldQuery=")
				})
			})

			for i := 0; i < len(entries); i++ {
				params := entries[i].Parameters[0].(listOpEntry)
				if len(params.queryTemplate) == 0 {
					panic("query templates missing")
				}
				var multiQueryValue string
				var queryValues []string

				fields := common.CopyObject(params.queryArgs)
				delete(fields, "labels")
				multiQueryValue, queryValues = expandFieldQuery(fields, params.queryTemplate)
				fquery := "fieldQuery" + "=" + multiQueryValue

				Context("with field query=", func() {
					for _, queryValue := range queryValues {
						query := "fieldQuery" + "=" + queryValue
						DescribeTable(fmt.Sprintf("%s", queryValue), func(test listOpEntry) {
							verifyListOp(test, query)
						}, entries[i])
					}

					if len(queryValues) > 1 {
						DescribeTable(fmt.Sprintf("%s", multiQueryValue), func(test listOpEntry) {
							verifyListOp(test, fquery)
						}, entries[i])
					}
				})

				labels := params.queryArgs["labels"]
				if t.SupportsLabels && labels != nil {

					multiQueryValue, queryValues = expandLabelQuery(labels.(map[string]interface{}), params.queryTemplate)
					lquery := "labelQuery" + "=" + multiQueryValue

					Context("with label query=", func() {
						for _, queryValue := range queryValues {
							query := "labelQuery" + "=" + queryValue
							DescribeTable(fmt.Sprintf("%s", queryValue), func(test listOpEntry) {
								verifyListOp(test, query)
							}, entries[i])
						}

						if len(queryValues) > 1 {
							DescribeTable(fmt.Sprintf("%s", multiQueryValue), func(test listOpEntry) {
								verifyListOp(test, lquery)
							}, entries[i])
						}
					})

					Context("with multiple field and label queries", func() {
						DescribeTable(fmt.Sprintf("%s", fquery+"&"+lquery), func(test listOpEntry) {
							verifyListOp(test, fquery+"&"+lquery)
						}, entries[i])
					})
				}
			}
		})
	})
}

func expandFieldQuery(fieldQueryArgs common.Object, queryTemplate string) (string, []string) {
	var expandedMultiQuery string
	var expandedQueries []string
	for k, v := range fieldQueryArgs {
		if v == nil {
			continue
		}

		if m, ok := v.(map[string]interface{}); ok {
			bytes, err := json.Marshal(m)
			if err != nil {
				panic(err)
			}
			v = string(bytes)
		}
		if a, ok := v.([]interface{}); ok {
			bytes, err := json.Marshal(a)
			if err != nil {
				panic(err)
			}
			v = string(bytes)

		}
		expandedQueries = append(expandedQueries, fmt.Sprintf(queryTemplate, k, v))
	}

	expandedMultiQuery = strings.Join(expandedQueries, "|")
	return expandedMultiQuery, expandedQueries
}

func expandLabelQuery(labelQueryArgs map[string]interface{}, queryTemplate string) (string, []string) {
	var expandedMultiQuery string
	var expandedQueries []string

	for key, values := range labelQueryArgs {
		for _, value := range values.([]interface{}) {
			expandedQueries = append(expandedQueries, fmt.Sprintf(queryTemplate, key, value))
		}
	}

	expandedMultiQuery = strings.Join(expandedQueries, "|")
	return expandedMultiQuery, expandedQueries
}
