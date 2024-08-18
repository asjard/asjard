package config

import (
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

const testYamlConent = `Hacker: true
name: steve
hobbies:
  - skateboarding
  - snowboarding
  - go
products:
- name: apple
  price: 1.23
- name: xx
  price: 4.56
clothing:
  jacket: leather
  trousers: denim
  pants:
      size: large
  products:
  - name: jacket
    price: 1.223
  - name: xx
    price: 4.56
age: 35
eyes : brown
beard: true`

const testJsonContent = `{
    "Hacker": true,
    "name": "steve",
    "hobbies": [
        "skateboarding",
        "snowboarding",
        "go"
    ],
    "products": [{
    	"name": "apple",
      	"price": 1.23
    },{
    	"name": "xx",
      	"price": 4.56
    }],
    "clothing": {
        "jacket": "leather",
        "trousers": "denim",
        "pants": {
            "size": "large"
        },
        "products": [{
        	"name": "jacket",
         	"price": 1.223
        },{
        	"name": "xx",
         	"price": 4.56
        }]
    },
    "age": 35,
    "eyes": "brown",
    "beard": true
}`

const testPropsContent = `
Hacker=true
name=steve
products[0].name=apple
products[0].price=1.23
products[1].name=xx
products[1].price=4.56
hobbies[0]=skateboarding
hobbies[1]=snowboarding
hobbies[2]=go
clothing.jacket=leather
clothing.trousers=denim
clothing.pants.size=large
clothing.products[0].name=jacket
clothing.products[0].price=1.223
clothing.products[1].name=xx
clothing.products[1].price=4.56
age=35
eyes=brown
beard=true
`

const testTomlContent = `
Hacker = true
name = "steve"
age = 35
eyes = "brown"
beard = true

hobbies = ["skateboarding", "snowboarding", "go"]

[[products]]
name = "apple"
price = 1.23

[[products]]
name = "xx"
price = 4.56

## products=[{"name": "apple", "price":1.23}, {"name":"xx","price":4.56}]

[clothing]
jacket = "leather"
trousers = "denim"

[clothing.pants]
size = "large"

[[clothing.products]]
name = "jacket"
price = 1.223

[[clothing.products]]
name = "xx"
price = 4.56
`

var expectPropsMap = map[string]any{
	"Hacker":                     true,
	"name":                       "steve",
	"hobbies[0]":                 "skateboarding",
	"hobbies[1]":                 "snowboarding",
	"hobbies[2]":                 "go",
	"products[0].name":           "apple",
	"products[0].price":          1.23,
	"products[1].name":           "xx",
	"products[1].price":          4.56,
	"clothing.jacket":            "leather",
	"clothing.trousers":          "denim",
	"clothing.pants.size":        "large",
	"clothing.products[0].name":  "jacket",
	"clothing.products[0].price": 1.223,
	"clothing.products[1].name":  "xx",
	"clothing.products[1].price": 4.56,
	"age":                        35,
	"eyes":                       "brown",
	"beard":                      true,
}

func TestConverToProperties(t *testing.T) {
	datas := []struct {
		ext     string
		content string
		expect  map[string]any
	}{
		{ext: "yaml", content: testYamlConent, expect: expectPropsMap},
		{ext: ".yaml", content: testYamlConent, expect: expectPropsMap},
		{ext: "yml", content: testYamlConent, expect: expectPropsMap},
		{ext: ".yml", content: testYamlConent, expect: expectPropsMap},
		{ext: "json", content: testJsonContent, expect: expectPropsMap},
		{ext: ".json", content: testJsonContent, expect: expectPropsMap},
		{ext: "props", content: testPropsContent, expect: expectPropsMap},
		{ext: "properties", content: testPropsContent, expect: expectPropsMap},
		{ext: ".props", content: testPropsContent, expect: expectPropsMap},
		{ext: ".properties", content: testPropsContent, expect: expectPropsMap},
		{ext: "toml", content: testTomlContent, expect: expectPropsMap},
		{ext: ".toml", content: testTomlContent, expect: expectPropsMap},
	}
	for _, data := range datas {
		propsMap, err := ConvertToProperties(data.ext, []byte(data.content))
		assert.Nil(t, err)
		for key, value := range expectPropsMap {
			v, ok := propsMap[key]
			if !ok {
				t.Errorf("test ext %s key %s not found", data.ext, key)
				t.FailNow()
			}
			if cast.ToString(v) != cast.ToString(value) {
				t.Errorf("test ext %s key %s value not equal, expect %v, actual %v",
					data.ext, key, value, v)
				t.FailNow()
			}
		}
	}
}
