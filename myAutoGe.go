package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type AddSource struct {
	CarID        uint32  `json:"car_id"`
	Price        int     `json:"price"`
	PriceUSD     float32 `json:"price_usd"`
	Currency     uint8   `json:"currency_id"`
	ManID        uint16  `json:"man_id"`
	ModelID      uint16  `json:"model_id"`
	CarModel     string  `json:"car_model"`
	ProdYear     uint16  `json:"prod_year"`
	EngineVolume uint16  `json:"engine_volume"`
	GearTypeID   uint16  `json:"gear_type_id"`
	VehicleType  uint16  `json:"vehicle_type"`
}

type Response struct {
	Data struct {
		AddSources []AddSource `json:"items"`
	} `json:"data"`
}

type Category struct {
	CategoryId uint16 `json:"category_id"`
	Name       string `json:"title"`
}

type Currency struct {
	ID     uint8  `json:"currencyID"`
	Name   string `json:"title"`
	Symbol string `json:"currencySymbol"`
}

type Model struct {
}

type Man struct {
	ID      uint16 `json:"man_id"`
	ManName string `json:"man_name"`
}

type Location struct {
}

type LoadedAppData struct {
	Categories []Category
	Currencies []Currency
	Models     []Model
	Mans       []Man
	Locations  []Location
}

type AppData struct {
	Categories map[uint16]Category
	Currencies map[uint16]Currency
	Models     map[uint16]Model
	Mans       map[uint16]Man
	Locations  map[uint16]Location
}

const url = "https://api2.myauto.ge/ka/products/"

var appData AppData

func MyAutoGeParsePage(page uint16) uint16 {
	loadData()

	addSources := loadPage(page)

	if len(addSources) == 0 {
		fmt.Println("0 Items - resetting page to 1")
	} else {
		fmt.Println(strconv.Itoa(len(addSources)) + " Items loaded")
		page++
	}

	carIds := make([]uint32, 0)

	for key, _ := range addSources {
		carIds = append(carIds, key)
	}

	RestoreTrashedAdds(carIds, "MyAutoGe")

	existingAdds := GetExistingAdds(carIds, "MyAutoGe")

	for id, add := range existingAdds {
		add.name = getName(addSources[id])
		//            'name' => $name,
		//            'description' => $this->getAddDescription($addSource),
		//            'price' => $addSource->price,
		//            'price_usd' => $addSource->price_usd,
		//            'currency' => $currency,
		//            'location_id' => $location->id,
		//            'category_id' => $category->id,
	}

	fmt.Println(existingAdds)

	return page
}

func getName(addSource AddSource) string {
	name := []string{}
	val, ok := appData.Mans[addSource.ManID]
	if ok {
		name = append(name, val.ManName)
	}
	//        $name = [];
	//        if (array_key_exists($addSource->man_id, $this->appdata->Mans)) {
	//            $name[] = $this->appdata->Mans[$addSource->man_id];
	//        }
	//        if (array_key_exists($addSource->model_id, $this->appdata->Models)) {
	//            $name[] = $this->appdata->Models[$addSource->model_id];
	//        }
	//        if ($addSource->car_model) {
	//            $name[] = $addSource->car_model;
	//        }
	//        if ($addSource->prod_year) {
	//            $name[] = $addSource->prod_year;
	//        }
	//        if ($addSource->engine_volume) {
	//            $transmission = '';
	//            switch ($this->getGearType($addSource->gear_type_id)) {
	//                case 'Automatic':
	//                case 'Tiptronic':
	//                    $transmission = 'AT';
	//                    break;
	//                case 'Manual':
	//                    $transmission = 'MT';
	//                    break;
	//                case 'Variator':
	//                    $transmission = 'CVT';
	//                    break;
	//            }
	//
	//            if ($addSource->vehicle_type === self::TYPE_MOTO) {
	//                $name[] = $addSource->engine_volume . 'cc';
	//            } else {
	//                $name[] = number_format($addSource->engine_volume / 1000, 1) . $transmission;
	//            }
	//        }
	//

	return strings.Join(name, " ")
}

func loadPage(page uint16) map[uint32]AddSource {
	params := map[string]string{
		"ForRent":       "0",
		"CurrencyID":    "1",
		"MileageType":   "1",
		"SortOrder":     "1",
		"Page":          strconv.Itoa(int(page)),
		"hideDealPrice": "1",
		"Locs":          "2.3.4.7.15.30.113.52.37.36.38.39.40.31.5.41.44.47.48.53.54.8.16.6.14.13.12.11.10.9.55.56.57.59.58.61.62.63.64.66.71.72.74.75.76.77.78.80.81.82.83.84.85.86.87.88.91.96.97.101.109",
	}

	body := LoadUrl(url, params)

	var responseObject Response
	err := json.Unmarshal(body, &responseObject)
	if err != nil {
		fmt.Printf("Error parsing JSON: %v\n", err)
		return nil
	}

	result := make(map[uint32]AddSource)

	for _, addSource := range responseObject.Data.AddSources {
		result[addSource.CarID] = addSource
	}

	return result
}

func loadData() {
	if len(appData.Categories) == 0 {
		body := LoadUrl(url+"/appdata/other_en.json", nil)

		var loadedAppData LoadedAppData

		err := json.Unmarshal(body, &loadedAppData)
		if err != nil {
			fmt.Printf("Error parsing JSON: %v\n", err)
			panic("Can't load myAutoGe appdata")
		}

		categories := make(map[uint16]Category)

		for _, category := range loadedAppData.Categories {
			categories[category.CategoryId] = category
		}

		appData = AppData{
			Categories: categories,
		}

		fmt.Println("Appdata loaded")
	}
}

//    private function loadData(): void
//    {
//        $this->appdata = Cache::remember('myauto_appdata_ru', 24 * 60 * 60, function() {
//            $result = $this->client->request('GET', self::URL . '/appdata/other_en.json');
//            $data = json_decode($result->getBody()->getContents(), false, 512, JSON_THROW_ON_ERROR);
//
//            $data->Locations->items = array_reduce($data->Locations->items, static function($carry, $item) {
//                $carry[$item->location_id] = $item;
//                return $carry;
//            }, []);
//
//            $data->Mans = array_reduce($data->Mans, static function($carry, $item) {
//                $carry[$item->man_id] = $item->man_name;
//                return $carry;
//            }, []);
//
//            $data->Models = array_reduce($data->Models, static function($carry, $item) {
//                $carry[$item->model_id] = $item->model_name;
//                return $carry;
//            }, []);
//
//            $data->Currencies = array_reduce($data->Currencies, static function($carry, $item) {
//                $carry[$item->currencyID] = $item->title;
//                return $carry;
//            }, []);
//
//            return $data;
//        });
//    }
