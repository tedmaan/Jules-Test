package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Глобальные переменные для подключения к MongoDB и коллекции.
var client *mongo.Client
var haikusCollection *mongo.Collection

// testMode глобальная переменная, которая будет установлена в
// true, если запуск с флагом -test.
var testMode bool

// Haiku представляет собой структуру документа, который будет
// храниться в MongoDB.
type Haiku struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"` // ID документа MongoDB.
	Text      string             `bson:"text"`          // Текст хайку.
	Timestamp time.Time          `bson:"timestamp"`     // Время генерации хайку.
	Moisture  int                `bson:"moisture"`      // Показатель влажности с датчика.
	Light     int                `bson:"light"`         // Показатель освещенности с датчика.
	Temp      int                `bson:"temperature"`   // Показатель температуры с датчика.
	PH        int                `bson:"ph"`            // Показатель рН с датчика.
}

// SensorData содержит текущие показания датчиков.
type SensorData struct {
	Moisture     int
	Illumination int
	Temperature  int
	PH           int
}

// LLMRequest представляет структуру запроса к API нейросети.
type LLMRequest struct {
	Model    string `json:"model"` // Идентификатор модели нейросети.
	Messages []struct {
		Role    string `json:"role"`    // Роль отправителя сообщения (например, "user").
		Content string `json:"content"` // Содержание сообщения (промпт).
	} `json:"messages"`
}

// LLMResponse представляет структуру ответа от API нейросети.
type LLMResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"` // Содержание ответа нейросети.
		} `json:"message"`
	} `json:"choices"`
}

// initMongoDB инициализирует клиент MongoDB и подключается к
// указанной коллекции.
func initMongoDB() {
	// !!! ЗАМЕНИТЕ ЭТУ ЗАГЛУШКУ НА ВАШ АДРЕС ПОДКЛЮЧЕНИЯ К MONGODB ATLAS!!!
	// Пример: "mongodb+srv://user:password@cluster0.abcde.mongodb.net/gardenDB?retryWrites=true&w=majority"
	mongoURI := "mongodb+srv://Jules:Str0ngJulesPwd@haiku0.rs7dhjr.mongodb.net/?retryWrites=true&w=majority&appName=Haiku0"
	dbName := "gardenHaikuDB"
	collectionName := "haikus" // Имя коллекции, где будут храниться хайку.

	var err error
	client, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Ошибка подключения к MongoDB: %v", err)
	}

	// Проверяем подключение, отправляя пинг-запрос к базе данных.
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Ошибка пинга MongoDB: %v", err)
	}

	log.Println("Успешное подключение к MongoDB!")
	haikusCollection = client.Database(dbName).Collection(collectionName)
}

// insertHaiku вставляет новое хайку в базу данных MongoDB.
func insertHaiku(haiku Haiku) error {
	_, err := haikusCollection.InsertOne(context.TODO(), haiku)
	if err != nil {
		return fmt.Errorf("ошибка вставки хайку: %w", err)
	}
	log.Printf("Хайку успешно сохранено: %s", haiku.Text)
	return nil
}

// getAllHaikus извлекает все хайку из базы данных, отсортированные
// по метке времени в убывающем порядке.
func getAllHaikus() ([]Haiku, error) {
	// Опция сортировки по времени в порядке убывания (самые новые сверху).
	findOptions := options.Find().SetSort(bson.D{{Key: "timestamp", Value: -1}})
	cursor, err := haikusCollection.Find(context.TODO(), bson.D{}, findOptions)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения хайку: %w", err)
	}
	defer cursor.Close(context.TODO()) // Закрываем курсор после завершения функции.

	var haikus []Haiku

	// Декодируем все документы из курсора в срез структур Haiku.
	if err = cursor.All(context.TODO(), &haikus); err != nil {
		return nil, fmt.Errorf("ошибка декодирования хайку: %w", err)
	}

	return haikus, nil
}

// simulateSensorData генерирует случайные данные датчиков для демонстрации.
func simulateSensorData() SensorData {
	log.Println("Использую имитированные данные датчиков (режим тестирования)...")
	return SensorData{
		Moisture:     rand.Intn(1024),  // Влажность: 0-1023
		Illumination: rand.Intn(1024),  // Освещенность: 0-1023
		Temperature:  rand.Intn(41),    // Температура: 0-40 Цельсия
		PH:           rand.Intn(15),    // pH: 0-14
	}
}

// readSensorDataFromRaspberryPi имитирует чтение данных с контактов Raspberry Pi.
// В реальном приложении здесь будет использоваться библиотека periph.io для
// взаимодействия с аппаратными датчиками, например, с помощью GPIO, I2C, SPI.
func init() {
	// Загружаем драйверы periph.io для обнаружения устройств.
	// if _, err := host.Init(); err != nil {
	// 	log.Fatal(err)
	// }
}

func readSensorDataFromRaspberryPi() SensorData {
	log.Println("Чтение данных с реальных контактов Raspberry Pi (имитация)...")
	// Для текущего примера пока возвращаем псевдореальные случайные данные,
	// чтобы продемонстрировать функционал без фактического оборудования.
	return SensorData{
		Moisture:     550 + rand.Intn(100),  // Некий "реалистичный" диапазон
		Illumination: 600 + rand.Intn(200),
		Temperature:  25 + rand.Intn(5),
		PH:           7 + rand.Intn(2),
	}
}

// buildLLMPrompt конструирует промпт для нейросети на основе данных датчиков.
func buildLLMPrompt(data SensorData) string {
	promptTemplate := `Make haiku reflecting these parameters:
if moisture between 0-200 reflect drought
if moisture between 201-400 reflect dryness
if moisture between 401-700 reflect normal moisture, thriving
if moisture between 701-900 reflect wetness, dew
if moisture between 901-1023 reflect oversaturation, puddles
if illumination between 0-200 reflect night, darkness
if illumination between 201-400 reflect dawn, early morning
if illumination between 401-700 reflect daylight, sunshine
if illumination between 701-1023 reflect bright sun, strong light
if temperature between 0-10 reflect cold, frost
if temperature between 11-20 reflect coolness, pleasant
if temperature between 21-30 reflect warmth, growth
if temperature between 31-40 reflect heat, summer
if pH between 0-6 reflect acidity, sourness, difficulty
if pH between 7 reflect neutrality, balance
if pH between 8-14 reflect alkalinity, bitterness, struggle

The parameters now are:
Moisture %d
Illumination %d
Temperature %d
pH %d

Separate haiku from other text with $ symbols like $(haiku)$`

	// Форматируем шаблон промпта, подставляя актуальные данные датчиков.
	return fmt.Sprintf(promptTemplate, data.Moisture,
		data.Illumination, data.Temperature, data.PH)
}

// callLLM отправляет промпт к API нейросети и возвращает её ответ.
func callLLM(prompt string) (string, error) {
	llmAPIHost := "https://llm.chutes.ai/v1/chat/completions"
	llmModelID := "deepseek-ai/DeepSeek-R1-0528"
	llmAPIKey := "cpk_b6594cef5d42450bbc31d99e3fb5e04f.1133333fbecd561aae8e4836dbff4b49.1U2TZHPOc0zffpvXZ15pXACypqvneNFx" // Ваш API ключ

	// Формируем тело запроса в формате JSON.
	reqBody := LLMRequest{
		Model: llmModelID,
		Messages: []struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		}{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody) // Сериализуем структуру запроса в JSON.
	if err != nil {
		return "", fmt.Errorf("ошибка сериализации запроса LLM: %w", err)
	}

	// Создаем новый HTTP POST-запрос.
	req, err := http.NewRequest("POST", llmAPIHost,
		bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("ошибка создания HTTP запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")      // Устанавливаем заголовок Content-Type.
	req.Header.Set("Authorization", "Bearer "+llmAPIKey) // Устанавливаем заголовок авторизации с API-ключом.

	client := &http.Client{Timeout: 30 * time.Second} // Создаем HTTP-клиент с таймаутом.
	resp, err := client.Do(req)                       // Отправляем запрос.
	if err != nil {
		return "", fmt.Errorf("ошибка отправки HTTP запроса к LLM: %w", err)
	}
	defer resp.Body.Close() // Гарантируем закрытие тела ответа.

	// Проверяем статус код ответа.
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM API вернул ошибку: %d %s",
			resp.StatusCode, string(bodyBytes))
	}

	var llmResp LLMResponse
	err = json.NewDecoder(resp.Body).Decode(&llmResp) // Декодируем JSON-ответ в структуру.
	if err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа LLM: %w", err)
	}

	// Проверяем, что ответ содержит содержимое.
	if len(llmResp.Choices) == 0 ||
		len(llmResp.Choices[0].Message.Content) == 0 {
		return "", fmt.Errorf("LLM не вернул контент")
	}

	return llmResp.Choices[0].Message.Content, nil
}

// extractHaiku извлекает текст хайку из полного ответа нейросети,
// используя заданный разделитель $(haiku)$.
func extractHaiku(llmOutput string) (string, error) {
	// Регулярное выражение для поиска текста внутри $(...)$.
	re := regexp.MustCompile(`\$\((.*?)\)\$`)
	matches := re.FindStringSubmatch(llmOutput)

	if len(matches) > 1 {
		return strings.TrimSpace(matches[1]), nil // Возвращаем извлеченный текст, удаляя пробелы.
	}

	return "", fmt.Errorf("хайку не найдено в ответе LLM, отсутствует разделитель $(haiku)$")
}

// sensorPoller это горутина, которая имитирует считывание данных датчиков,
// вызывает нейросеть для генерации хайку и сохраняет его в базе данных.
func sensorPoller() {
	// Определяем, какую функцию для получения данных датчиков использовать.
	var getSensorData func() SensorData
	if testMode {
		getSensorData = simulateSensorData
		log.Println("Запущен в режиме тестирования. Будут использоваться имитированные данные датчиков.")
	} else {
		getSensorData = readSensorDataFromRaspberryPi
		log.Println("Запущен в обычном режиме. Будут использоваться (имитированные пока) данные с контактов Raspberry Pi.")
	}

	// Выполняем первый замер сразу после запуска, чтобы страница не
	// была пустой при первом открытии.
	log.Println("Выполняю первый замер данных датчиков и генерирую хайку при запуске...")
	data := getSensorData() // Используем выбранную функцию
	log.Printf("Сырые данные датчиков: Влажность=%d, Освещенность=%d, Температура=%d, pH=%d",
		data.Moisture, data.Illumination, data.Temperature, data.PH)

	prompt := buildLLMPrompt(data)
	log.Printf("Сгенерированный промпт для LLM:\n%s", prompt)

	llmResponse, err := callLLM(prompt)
	if err != nil {
		log.Printf("Ошибка вызова LLM: %v", err)
	} else {
		log.Printf("Полный ответ LLM:\n%s", llmResponse)
		haikuText, err := extractHaiku(llmResponse)
		if err != nil {
			log.Printf("Ошибка извлечения хайку из ответа LLM: %v", err)
		} else {
			newHaiku := Haiku{
				Text:      haikuText,
				Timestamp: time.Now(),
				Moisture:  data.Moisture,
				Light:     data.Illumination,
				Temp:      data.Temperature,
				PH:        data.PH,
			}
			err = insertHaiku(newHaiku)
			if err != nil {
				log.Printf("Ошибка сохранения хайку в базе данных: %v", err)
			}
		}
	}

	// Запускаем таймер, который будет срабатывать раз в час.
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop() // Гарантируем остановку таймера при выходе из функции.

	for range ticker.C { // Цикл будет выполняться каждый раз, когда срабатывает таймер.
		log.Println("Выполняю замер данных датчиков и генерирую хайку...")
		data := getSensorData() // Получаем данные датчиков с помощью выбранной функции.
		log.Printf("Сырые данные датчиков: Влажность=%d, Освещенность=%d, Температура=%d, pH=%d",
			data.Moisture, data.Illumination, data.Temperature, data.PH)

		prompt = buildLLMPrompt(data) // Строим промпт для нейросети.
		log.Printf("Сгенерированный промпт для LLM:\n%s", prompt)

		llmResponse, err = callLLM(prompt) // Вызываем API нейросети.
		if err != nil {
			log.Printf("Ошибка вызова LLM: %v", err)
			continue // Продолжаем к следующему циклу, если произошла ошибка.
		}

		log.Printf("Полный ответ LLM: \n%s", llmResponse)
		haikuText, err := extractHaiku(llmResponse) // Извлекаем хайку из ответа.
		if err != nil {
			log.Printf("Ошибка извлечения хайку из ответа LLM: %v", err)
			continue // Продолжаем к следующему циклу.
		}

		// Создаем новую структуру Haiku для сохранения.
		newHaiku := Haiku{
			Text:      haikuText,
			Timestamp: time.Now(),
			Moisture:  data.Moisture,
			Light:     data.Illumination,
			Temp:      data.Temperature,
			PH:        data.PH,
		}

		err = insertHaiku(newHaiku) // Сохраняем хайку в MongoDB.
		if err != nil {
			log.Printf("Ошибка сохранения хайку в базе данных: %v", err)
		}
	}
}

// haikuHandler обрабатывает HTTP-запросы к корневому пути,
// извлекает все хайку из базы данных и отображает их на HTML-странице.
func haikuHandler(w http.ResponseWriter, r *http.Request) {
	haikus, err := getAllHaikus() // Получаем все хайку из MongoDB.
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка получения хайку из базы данных: %v", err), http.StatusInternalServerError)
		return
	}

	// HTML-шаблон для отображения хайку.
	// Используется Tailwind CSS для стилизации и шрифт Inter.
	tmpl := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Garden Haikus</title>
    <link href="https://cdn.jsdelivr.net/npm/tailwindcss@2.2.19/dist/tailwind.min.css" rel="stylesheet">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;700&display=swap" rel="stylesheet">
    <style>
        body { font-family: 'Inter', sans-serif; background-color: #f0fdf4; color: #166534; }
        .haiku-card { background-color: #dcfce7; border: 1px solid #a7f3d0; box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1); }
        .haiku-card p { white-space: pre-wrap; } /* Сохраняет переносы строк хайку */
    </style>
</head>
<body class="p-8">
    <div class="max-w-4xl mx-auto">
        <h1 class="text-4xl font-bold text-center mb-8 text-green-800">Garden Haikus 🌿</h1>
        <div class="space-y-6">
            {{ if . }}
                {{ range . }}
                <div class="haiku-card p-6 rounded-lg">
                    <p class="text-lg font-medium text-green-700 mb-2">{{ .Text }}</p>
                    <p class="text-sm text-green-600">
                        <span class="font-semibold">Generated At:</span> {{ .Timestamp.Format "2006-01-02 15:04:05" }}<br>
                        <span class="font-semibold">Moisture:</span> {{ .Moisture }} |
                        <span class="font-semibold">Illumination:</span> {{ .Light }} |
                        <span class="font-semibold">Temperature:</span> {{ .Temp }} |
                        <span class="font-semibold">pH:</span> {{ .PH }}
                    </p>
                </div>
                {{ end }}
            {{ else }}
                <p class="text-center text-xl text-gray-500">No haikus generated yet. Please wait for the next hourly sensor reading.</p>
            {{ end }}
        </div>
    </div>
</body>
</html>`

	t, err := template.New("haikuPage").Parse(tmpl) // Создаем и парсим HTML-шаблон.
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка парсинга HTML шаблона: %v", err), http.StatusInternalServerError)
		return
	}

	// Выполняем шаблон, передавая данные хайку.
	err = t.Execute(w, haikus)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка выполнения HTML шаблона: %v", err), http.StatusInternalServerError)
	}
}

func main() {
	// Инициализируем флаг -test.
	flag.BoolVar(&testMode, "test", false, "Use simulated sensor data instead of Raspberry Pi pins")
	// Парсим аргументы командной строки.
	flag.Parse()

	// Инициализируем генератор случайных чисел.
	rand.Seed(time.Now().UnixNano())

	// Инициализируем подключение к MongoDB.
	initMongoDB()

	// Отложенное отключение от MongoDB при завершении работы программы.
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatalf("Ошибка отключения от MongoDB: %v", err)
		}
		log.Println("Отключение от MongoDB.")
	}()

	// Запускаем горутину, которая будет периодически опрашивать датчики
	// и генерировать хайку.
	go sensorPoller()

	// Настраиваем HTTP-сервер.
	// Все запросы к корневому пути будут обрабатываться функцией haikuHandler.
	http.HandleFunc("/", haikuHandler)
	log.Println("Веб-сервер запущен на http://localhost:8080")

	// Запускаем сервер и блокируем выполнение main-функции до тех пор,
	// пока сервер не завершится (например, из-за ошибки).
	log.Fatal(http.ListenAndServe(":8080", nil))
}
