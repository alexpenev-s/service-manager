/*
 *    Copyright 2018 The Service Manager Authors
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */
package broker_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/Peripli/service-manager/test/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cast"
)

func TestBrokers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Broker API Tests Suite")
}

var _ = Describe("Service Manager Broker API", func() {

	var (
		ctx          *common.TestContext
		brokerServer *common.BrokerServer

		brokerServerJSON       common.Object
		expectedBrokerResponse common.Object
	)

	BeforeSuite(func() {
		brokerServer = common.NewBrokerServer()
		ctx = common.NewTestContext(nil)
	})

	AfterSuite(func() {
		ctx.Cleanup()
		if brokerServer != nil {
			brokerServer.Close()
		}
	})

	BeforeEach(func() {
		brokerServer.Reset()
		brokerName := "brokerName"
		brokerDescription := "description"

		brokerServerJSON = common.Object{
			"name":        brokerName,
			"broker_url":  brokerServer.URL,
			"description": brokerDescription,
			"credentials": common.Object{
				"basic": common.Object{
					"username": brokerServer.Username,
					"password": brokerServer.Password,
				},
			},
		}
		expectedBrokerResponse = common.Object{
			"name":        brokerName,
			"broker_url":  brokerServer.URL,
			"description": brokerDescription,
		}
		common.RemoveAllBrokers(ctx.SMWithOAuth)
	})

	Describe("GET", func() {
		var id string

		AfterEach(func() {
			assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
		})

		Context("when the broker does not exist", func() {
			It("returns 404", func() {
				ctx.SMWithOAuth.GET("/v1/service_brokers/12345").
					Expect().
					Status(http.StatusNotFound).
					JSON().Object().
					Keys().Contains("error", "description")
			})
		})

		Context("when the broker exists", func() {
			BeforeEach(func() {
				reply := ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
					Expect().
					Status(http.StatusCreated).
					JSON().Object().
					ContainsMap(expectedBrokerResponse)

				id = reply.Value("id").String().Raw()

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				brokerServer.ResetCallHistory()
			})

			It("returns the broker with given id", func() {
				ctx.SMWithOAuth.GET("/v1/service_brokers/"+id).
					Expect().
					Status(http.StatusOK).
					JSON().Object().
					ContainsMap(expectedBrokerResponse).
					Keys().NotContains("credentials", "services")
			})
		})
	})

	Describe("List", func() {
		AfterEach(func() {
			assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
		})

		Context("when no brokers exist", func() {
			It("returns empty array", func() {
				ctx.SMWithOAuth.GET("/v1/service_brokers").
					Expect().
					Status(http.StatusOK).
					JSON().Object().Value("brokers").Array().
					Empty()
			})
		})

		Context("when brokers exist", func() {
			BeforeEach(func() {
				ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
					Expect().
					Status(http.StatusCreated).
					JSON().Object().
					ContainsMap(expectedBrokerResponse).
					Keys().
					NotContains("credentials", "services")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				brokerServer.ResetCallHistory()
			})

			It("returns all without catalog if no query parameter is provided", func() {
				ctx.SMWithOAuth.GET("/v1/service_brokers").
					Expect().
					Status(http.StatusOK).
					JSON().Object().Value("brokers").Array().First().Object().
					ContainsMap(expectedBrokerResponse).
					Keys().
					NotContains("credentials", "services")
			})

			It("returns all with catalog if query parameter is provided", func() {
				ctx.SMWithOAuth.GET("/v1/service_brokers").WithQuery("catalog", true).
					Expect().
					Status(http.StatusOK).
					JSON().Object().Value("brokers").Array().First().Object().
					ContainsMap(expectedBrokerResponse).
					ContainsKey("services").
					NotContainsKey("credentials")
			})
		})
	})

	Describe("POST", func() {
		Context("when content type is not JSON", func() {
			It("returns 415", func() {
				ctx.SMWithOAuth.POST("/v1/service_brokers").WithText("text").
					Expect().
					Status(http.StatusUnsupportedMediaType).
					JSON().Object().
					Keys().Contains("error", "description")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
			})
		})

		Context("when request body is not a valid JSON", func() {
			It("returns 400", func() {
				ctx.SMWithOAuth.POST("/v1/service_brokers").
					WithText("invalid json").
					WithHeader("content-type", "application/json").
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().
					Keys().Contains("error", "description")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
			})
		})

		Context("when a request body field is missing", func() {
			assertPOSTReturns400WhenFieldIsMissing := func(field string) {
				BeforeEach(func() {
					delete(brokerServerJSON, field)
					delete(expectedBrokerResponse, field)
				})

				It("returns 400", func() {
					ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
						Expect().
						Status(http.StatusBadRequest).
						JSON().Object().
						Keys().Contains("error", "description")

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
				})
			}

			assertPOSTReturns201WhenFieldIsMissing := func(field string) {
				BeforeEach(func() {
					delete(brokerServerJSON, field)
					delete(expectedBrokerResponse, field)
				})

				It("returns 201", func() {
					ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
						Expect().
						Status(http.StatusCreated).
						JSON().Object().
						ContainsMap(expectedBrokerResponse).
						Keys().NotContains("services", "credentials")

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				})
			}

			Context("when name field is missing", func() {
				assertPOSTReturns400WhenFieldIsMissing("name")
			})

			Context("when broker_url field is missing", func() {
				assertPOSTReturns400WhenFieldIsMissing("broker_url")
			})

			Context("when credentials field is missing", func() {
				assertPOSTReturns400WhenFieldIsMissing("credentials")
			})

			Context("when description field is missing", func() {
				assertPOSTReturns201WhenFieldIsMissing("description")
			})

		})

		Context("when fetching catalog fails", func() {
			BeforeEach(func() {
				brokerServer.CatalogHandler = func(w http.ResponseWriter, req *http.Request) {
					common.SetResponse(w, http.StatusInternalServerError, common.Object{})
				}
			})

			It("returns an error", func() {
				ctx.SMWithOAuth.POST("/v1/service_brokers").
					WithJSON(brokerServerJSON).
					Expect().Status(http.StatusInternalServerError).
					JSON().Object().
					Keys().Contains("error", "description")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
			})
		})

		Context("when request is successful", func() {
			assertPOSTReturns201 := func() {
				It("returns 201", func() {
					ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
						Expect().
						Status(http.StatusCreated).
						JSON().Object().
						ContainsMap(expectedBrokerResponse).
						Keys().NotContains("services", "credentials")

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				})
			}

			Context("when broker URL does not end with trailing slash", func() {
				BeforeEach(func() {
					brokerServerJSON["broker_url"] = strings.TrimRight(cast.ToString(brokerServerJSON["broker_url"]), "/")
					expectedBrokerResponse["broker_url"] = strings.TrimRight(cast.ToString(expectedBrokerResponse["broker_url"]), "/")
				})

				assertPOSTReturns201()
			})

			Context("when broker URL ends with trailing slash", func() {
				BeforeEach(func() {
					brokerServerJSON["broker_url"] = cast.ToString(brokerServerJSON["broker_url"]) + "/"
					expectedBrokerResponse["broker_url"] = cast.ToString(expectedBrokerResponse["broker_url"]) + "/"
				})

				assertPOSTReturns201()
			})
		})

		Context("when broker with name already exists", func() {
			It("returns 409", func() {
				ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
					Expect().
					Status(http.StatusCreated)

				ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
					Expect().
					Status(http.StatusConflict).
					JSON().Object().
					Keys().Contains("error", "description")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 2)
			})
		})
	})

	Describe("PATCH", func() {
		var id string

		BeforeEach(func() {
			reply := ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
				Expect().
				Status(http.StatusCreated).
				JSON().Object().
				ContainsMap(expectedBrokerResponse)

			id = reply.Value("id").String().Raw()

			assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
			brokerServer.ResetCallHistory()
		})

		Context("when content type is not JSON", func() {
			It("returns 415", func() {
				ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
					WithText("text").
					Expect().Status(http.StatusUnsupportedMediaType).
					JSON().Object().
					Keys().Contains("error", "description")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
			})
		})

		Context("when broker is missing", func() {
			It("returns 404", func() {
				ctx.SMWithOAuth.PATCH("/v1/service_brokers/no_such_id").
					WithJSON(brokerServerJSON).
					Expect().Status(http.StatusNotFound).
					JSON().Object().
					Keys().Contains("error", "description")
			})
		})

		Context("when request body is not valid JSON", func() {
			It("returns 400", func() {
				ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
					WithText("invalid json").
					WithHeader("content-type", "application/json").
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().
					Keys().Contains("error", "description")
			})
		})

		Context("when request body contains invalid credentials", func() {
			It("returns 400", func() {
				ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
					WithJSON(common.Object{"credentials": "123"}).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().
					Keys().Contains("error", "description")
			})
		})

		Context("when request body contains incomplete credentials", func() {
			It("returns 400", func() {
				ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
					WithJSON(common.Object{"credentials": common.Object{"basic": common.Object{"password": ""}}}).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().
					Keys().Contains("error", "description")
			})
		})

		Context("when broker with the name already exists", func() {
			var anotherTestBroker common.Object
			var anotherBrokerServer *common.BrokerServer

			BeforeEach(func() {
				anotherBrokerServer = common.NewBrokerServer()
				anotherBrokerServer.Username = "username"
				anotherBrokerServer.Password = "password"
				anotherTestBroker = common.Object{
					"name":        "another_name",
					"broker_url":  anotherBrokerServer.URL,
					"description": "another_description",
					"credentials": common.Object{
						"basic": common.Object{
							"username": anotherBrokerServer.Username,
							"password": anotherBrokerServer.Password,
						},
					},
				}
			})

			AfterEach(func() {
				if anotherBrokerServer != nil {
					anotherBrokerServer.Close()
				}
			})

			FIt("returns 409", func() {
				ctx.SMWithOAuth.POST("/v1/service_brokers").
					WithJSON(anotherTestBroker).
					Expect().
					Status(http.StatusCreated)

				assertInvocationCount(anotherBrokerServer.CatalogEndpointRequests, 1)

				ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
					WithJSON(anotherTestBroker).
					Expect().Status(http.StatusConflict).
					JSON().Object().
					Keys().Contains("error", "description")

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
			})
		})

		Context("when credentials are updated", func() {
			It("returns 200", func() {
				brokerServer.Username = "updatedUsername"
				brokerServer.Password = "updatedPassword"
				updatedCredentials := common.Object{
					"credentials": common.Object{
						"basic": common.Object{
							"username": brokerServer.Username,
							"password": brokerServer.Password,
						},
					},
				}
				reply := ctx.SMWithOAuth.PATCH("/v1/service_brokers/" + id).
					WithJSON(updatedCredentials).
					Expect().
					Status(http.StatusOK).
					JSON().Object()

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)

				reply = ctx.SMWithOAuth.GET("/v1/service_brokers/" + id).
					Expect().
					Status(http.StatusOK).
					JSON().Object()
				reply.ContainsMap(expectedBrokerResponse)
			})
		})

		Context("when created_at provided in body", func() {
			It("should not change created_at", func() {
				createdAt := "2015-01-01T00:00:00Z"

				ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
					WithJSON(common.Object{"created_at": createdAt}).
					Expect().
					Status(http.StatusOK).JSON().Object().
					ContainsKey("created_at").
					ValueNotEqual("created_at", createdAt)

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)

				ctx.SMWithOAuth.GET("/v1/service_brokers/"+id).
					Expect().
					Status(http.StatusOK).JSON().Object().
					ContainsKey("created_at").
					ValueNotEqual("created_at", createdAt)
			})
		})

		Context("when new broker server is available", func() {
			var (
				updatedBrokerServer           *common.BrokerServer
				updatedBrokerJSON             common.Object
				expectedUpdatedBrokerResponse common.Object
			)

			BeforeEach(func() {
				updatedBrokerServer = common.NewBrokerServer()
				updatedBrokerServer.Username = "updated_user"
				updatedBrokerServer.Password = "updated_password"
				updatedBrokerJSON = common.Object{
					"name":        "updated_name",
					"description": "updated_description",
					"broker_url":  updatedBrokerServer.URL,
					"credentials": common.Object{
						"basic": common.Object{
							"username": updatedBrokerServer.Username,
							"password": updatedBrokerServer.Password,
						},
					},
				}

				expectedUpdatedBrokerResponse = common.Object{
					"name":        updatedBrokerJSON["name"],
					"description": updatedBrokerJSON["description"],
					"broker_url":  updatedBrokerJSON["broker_url"],
				}
			})

			AfterEach(func() {
				if updatedBrokerServer != nil {
					updatedBrokerServer.Close()
				}
			})

			Context("when all updatable fields are updated at once", func() {
				It("returns 200", func() {
					ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
						WithJSON(updatedBrokerJSON).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(expectedUpdatedBrokerResponse).
						Keys().NotContains("services", "credentials")

					assertInvocationCount(updatedBrokerServer.CatalogEndpointRequests, 1)

					ctx.SMWithOAuth.GET("/v1/service_brokers/"+id).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(expectedUpdatedBrokerResponse).
						Keys().NotContains("services", "credentials")
				})
			})

			Context("when broker_url is changed and the credentials are correct", func() {
				It("returns 200", func() {
					updatedBrokerJSON := common.Object{
						"broker_url": updatedBrokerServer.URL,
					}
					updatedBrokerServer.Username = brokerServer.Username
					updatedBrokerServer.Password = brokerServer.Password

					ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
						WithJSON(updatedBrokerJSON).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(updatedBrokerJSON).
						Keys().NotContains("services", "credentials")

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
					assertInvocationCount(updatedBrokerServer.CatalogEndpointRequests, 1)

					ctx.SMWithOAuth.GET("/v1/service_brokers/"+id).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(updatedBrokerJSON).
						Keys().NotContains("services", "credentials")
				})
			})

			Context("when broker_url is changed but the credentials are wrong", func() {
				It("returns 500", func() {
					updatedBrokerJSON := common.Object{
						"broker_url": updatedBrokerServer.URL,
					}
					ctx.SMWithOAuth.PATCH("/v1/service_brokers/" + id).
						WithJSON(updatedBrokerJSON).
						Expect().
						Status(http.StatusInternalServerError)

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)

					ctx.SMWithOAuth.GET("/v1/service_brokers/"+id).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(expectedBrokerResponse).
						Keys().NotContains("services", "credentials")
				})
			})

		})

		for _, prop := range []string{"name", "description"} {
			Context("when only '"+prop+"' is updated", func() {
				It("returns 200", func() {
					updatedBrokerJSON := common.Object{}
					updatedBrokerJSON[prop] = "updated"
					ctx.SMWithOAuth.PATCH("/v1/service_brokers/"+id).
						WithJSON(updatedBrokerJSON).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(updatedBrokerJSON).
						Keys().NotContains("services", "credentials")

					ctx.SMWithOAuth.GET("/v1/service_brokers/"+id).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						ContainsMap(updatedBrokerJSON).
						Keys().NotContains("services", "credentials")

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				})
			})
		}

		Context("when not updatable fields are provided in the request body", func() {
			Context("when broker id is provided in request body", func() {
				It("should not create the broker", func() {
					brokerServerJSON = common.Object{"id": "123"}
					ctx.SMWithOAuth.PATCH("/v1/service_brokers/" + id).
						WithJSON(brokerServerJSON).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						NotContainsMap(brokerServerJSON)

					ctx.SMWithOAuth.GET("/v1/service_brokers/123").
						Expect().
						Status(http.StatusNotFound)

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				})
			})

			Context("when unmodifiable fields are provided in the request body", func() {
				BeforeEach(func() {
					brokerServerJSON = common.Object{
						"created_at": "2016-06-08T16:41:26Z",
						"updated_at": "2016-06-08T16:41:26Z",
						"services":   common.Array{common.Object{"name": "serviceName"}},
					}
				})

				It("should not change them", func() {
					ctx.SMWithOAuth.PATCH("/v1/service_brokers/" + id).
						WithJSON(brokerServerJSON).
						Expect().
						Status(http.StatusOK).
						JSON().Object().
						NotContainsMap(brokerServerJSON)

					ctx.SMWithOAuth.GET("/v1/service_brokers").
						Expect().
						Status(http.StatusOK).
						JSON().Object().Value("brokers").Array().First().Object().
						ContainsMap(expectedBrokerResponse)

					assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				})
			})
		})

		Context("when underlying broker catalog is modified", func() {
			BeforeEach(func() {
				brokerServer.Catalog = common.Object{
					"services": []interface{}{},
				}
			})
			//TODO more tests
			// when a new service is added , do patch, vetify it is retunrned by services api
			// when a new plan is added do patch, verify it is returned by plans api
			// when a service is removed
			// when a plan is removed
			// when a service's properties are modified
			// when a plan's properties are modified
			// fetch with catalog=true contains all known fields from the broker catalog - test how?
			//It("updates the catalog for the broker", func() {
			//	ctx.SMWithOAuth.PATCH("/v1/service_brokers/" + id).
			//		WithJSON(common.Object{}).
			//		Expect().
			//		Status(http.StatusOK)
			//
			//	//TODO subset the response catalog to the broker catalog (sm catalog has more fields..)
			//	ctx.SMWithOAuth.GET("/v1/service_brokers").
			//		WithQuery("catalog", true).
			//		Expect().
			//		Status(http.StatusOK).
			//		JSON().Object().Value("brokers").Array().First().Object().
			//		ContainsMap(expectedBrokerResponse).
			//		ContainsMap(brokerServer.Catalog)
			//
			//	assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
			//})
		})
	})

	Describe("DELETE", func() {
		AfterEach(func() {
			assertInvocationCount(brokerServer.CatalogEndpointRequests, 0)
		})

		Context("when broker does not exist", func() {
			It("returns 404", func() {
				ctx.SMWithOAuth.DELETE("/v1/service_brokers/999").
					Expect().
					Status(http.StatusNotFound).
					JSON().Object().
					Keys().Contains("error", "description")
			})
		})

		Context("when broker exists", func() {
			var id string

			BeforeEach(func() {
				reply := ctx.SMWithOAuth.POST("/v1/service_brokers").WithJSON(brokerServerJSON).
					Expect().
					Status(http.StatusCreated).
					JSON().Object().
					ContainsMap(expectedBrokerResponse)

				id = reply.Value("id").String().Raw()

				assertInvocationCount(brokerServer.CatalogEndpointRequests, 1)
				brokerServer.ResetCallHistory()
			})

			It("returns 200", func() {
				ctx.SMWithOAuth.GET("/v1/service_brokers/" + id).
					Expect().
					Status(http.StatusOK)

				ctx.SMWithOAuth.DELETE("/v1/service_brokers/" + id).
					Expect().
					Status(http.StatusOK).JSON().Object().Empty()

				ctx.SMWithOAuth.GET("/v1/service_brokers/" + id).
					Expect().
					Status(http.StatusNotFound)
			})

			It("deletes the related services and plans", func() {
				//TODO
			})
		})
	})
})

func assertInvocationCount(requests []*http.Request, invocationCount int) {
	Expect(len(requests)).To(Equal(invocationCount))
}
