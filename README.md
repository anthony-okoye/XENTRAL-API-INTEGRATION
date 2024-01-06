## **BACKEND**

Prerequisite:
_install GoLang locally_
_install PostgreSQL server locally_
_if changed from default - add variables for DB to .envrc file_

This is a simple CRUD backend service which attaches to the PostgreSQL DB and wraps simple operations including:

1. DB modeling and initialization (on startup)
2. Create operation (add row to table in DB, from given data)
3. Read operation (retrieve 1 row from database, from given ID)
4. List operation (retrieve X rows from database, from given metadata)
5. 1. List operation with relationships
6. Update operation (update 1 row in database, from given data)
7. Delete operation (soft-delete 1 row from database, from given ID)

Also, this service offers authentication services:

1. Create access/refresh token pairs
2. Use tokens in pairs to login, set through headers
3. Log the user out, if needed (destroy tokens)

## **_Explanations_**

_see Response section for universal response_

1. DB modeling and initialization (on startup)

-> First, to retrieve all libs, run in root folder "go mod tidy"
-> Run the service from _/cmd/server_ using "go run main.go"
(thats it, the service should be available for instance on https://localhost:8000 - this is the assumed URL from now on)

2. Create operation (add row to table in DB, from given data)
   _see Relations section for relation definition_

-> Execute _POST_ request:
Address: https://localhost:8000/create
Body:

```
{
    "entity": "category",           //required for entity determination
    "data": {                       //contains values for all fields
        "name": "testCategory",
        "active": true,
        "text": "testText"
                                    //...etc. (see Entities section for model specification)
    }
}
```

entities :
"product":
"review":
"order":
"user":
"discount":
"category":
"sales_channel
"cart"

3. Read operation (retrieve 1 row from database, from given ID)

-> Execute _POST_ request:
Address: https://localhost:8000/read
Body:

```
{
    "entity": "category",        //required for entity determination
    "data": {
        "id": 3                     //specify ID of entity to be read
    }
}
```

4. List operation (retrieve X rows from database, from given metadata)

-> Execute _POST_ request:
Address: https://localhost:8000/list
Body:

```
{
    "entity": "category",           //required for entity determination
    "metadata": {
        "filter": {
            "must": [               //hard filters, connected with AND logical operator (all apply)
                {
                    "key":   "name",
                    "value": "testCategory",
                    "type":  "eq",  //see Filters section for filter types
			    },
                {
                    "key":   "active",
                    "value": "true",
                    "type":  "eq",  //see Filters section for filter types
			    },
            ],
            "should": [              //soft filters, connected with OR logical operator (any one applies)
                // same as above
            ]
        },
        "orderBy": {                 //designate ordering of data in response
			"key":  "id",
			"type": "desc",          //asc - ascending, desc - descending
		},
        "fields": [                  //specify response fields
                                     //if empty retrieves all
                                     //will always retrieve non null fields: 'ID, CreatedAt, UpdatedAt, DeletedAt'
            "id",
            "name"
        ],
        "limit": 10,                //specify limit for DB retrieval (default 10, max 100)
        "page": 1                   //specify page for DB retrieval (default 0)
    }
}
```

4. 1. List operation with relationships (many 2 many relationships)
      -> Add relationships in list requests in order to retrieve related data, as well as the primary entities.
      Relationship names match the plural names of fields inside the data models.
      There are 3 ways of retrieving relation data:



## **XENTRAL-INTEGRATION**
Technical Documentation for Connecting with Xentral

1. POST Import orders request
https://ORGANISATION-ID.xentral.biz/api/salesOrders/actions/import


-> Request Headers
- Cache-Control : no-cache
- Postman-Token : <calculated when request is sent>
- Content-Length : 0
- Host : <calculated when request is sent>
- User-Agent : PostmanRuntime/7.32.1
- Accept : */*
- Accept-Encoding : gzip, deflate, br
- Connection : keep-alive
- Authorization : Bearer TOKEN
- Content-Type : application/vnd.xentral.default.v1-beta+json



2. POST create products
https://ORGANISATION-ID.xentral.biz/api/products


-> Request Headers
- Cache-Control : no-cache
- Postman-Token : <calculated when request is sent>
- Content-Type : application/json
- Content-Length : <calculated when request is sent>
- Host : <calculated when request is sent>
- User-Agent : PostmanRuntime/7.32.1
- Accept : */*
- Accept-Encoding : gzip, deflate, br
- Connection : keep-alive
- authorization : Bearer TOKEN
- content-type : application/vnd.xentral.default.v1+json

Body raw (json)
```
{
        "project": {"id": "1"},
        "measurements": {
            "width": {"unit": "cm", "value": 1.00},
            "height": {"unit": "cm", "value": 1.00},
            "length": {"unit": "cm", "value": 1.00},
            "weight": {"unit": "kg", "value": 1.00},
            "netWeight": {"unit": "kg", "value": 1.00}
        },
        "name": "First Bite2",
        "number": "31199715",
        "ean": "9781498934695",
        "shopPriceDisplay": "3.00",
        "description": "no value",
        "manufacturer": {
            "name": "Knox, Lorelei",
            "number": "no value",
            "link": "https://no_value"
          },
        "isStockItem": true,
        "minimumOrderQuantity": 1
    }
```
    
3. POST create sales order
https://ORGANISATION-ID.xentral.biz/api/salesOrders/actions/import


-> Request Headers
- authorization : Bearer TOKEN
- content-type : application/vnd.xentral.default.v1-beta+json
Body raw (json)
```
{
  "customer": {
    "id": "3"
  },
  "project": {
    "id": "1"
  },
  "financials": {
    "paymentMethod": {
      "id": "1"
    },
    "billingAddress": {
      "street": "Musterstraße 1",
      "country": "CH",
      "name": "test test",
      "city": "test local",
      "zipCode": "test local",
      "type": "mr"
    },
    "currency": "CHF"
  },
  "delivery": {
    "shippingAddress": {
      "street": "Musterstraße 1",
      "type": "mr",
      "name": "test test",
      "zipCode": "test local",
      "city": "Musterstadt",
      "country": "CH"
    },
    "shippingMethod": {
      "id": "1"
    }
  },
  "date": "12/11/2023",
  "positions": [
    {
      "product": {
        "id": "2"
      },
      "price": {
        "amount": "6.9",
        "currency": "CHF"
      },
      "quantity": 1
    }
  ]
}
```

4. GET list customers
https://ORGANISATION-ID.xentral.biz/api/customers?filter[0][key]=email&filter[0][value]=%s.com&filter[0][op]=equals


-> Request Headers
- authorization : Bearer TOKEN
- accept : application/vnd.xentral.default.v1+json

-> Query Params

- filter[0][key] : email
- filter[0][value] : EMAIL VALUE
- filter[0][op] : equals


4. GET list products
https://ORGANISATION-ID.xentral.biz/api/products?filter[0][key]=number&filter[0][value]=%s&filter[0][op]=equals


-> Request Headers
- authorization : Bearer TOKEN
- accept : application/vnd.xentral.default.v1+json

-> Query Params

- filter[0][key] : number
- filter[0][value] : NUMBER VALUE
- filter[0][op] : equals
