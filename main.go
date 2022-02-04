package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	MongoURI   string `envconfig:"MONGODB_URI" required:"true"`
	DBName     string `envconfig:"MONGODB_DBNAME" required:"true"`
	MongoMICol string `envconfig:"MONGODB_MICOL" required:"true"`
	MongoAgCol string `envconfig:"MONGODB_AGCOL" required:"true"`
}

type Meta struct {
	NoLoginRequired   bool   `json:"no_login_required,omitempty" bson:"no_login_required,omitempty"`
	NoCaptchaRequired bool   `json:"no_captcha_required,omitempty" bson:"no_captcha_required,omitempty"`
	Access            string `json:"access,omitempty" bson:"access,omitempty"`
	Extension         string `json:"extension,omitempty" bson:"extension,omitempty"`
	StrictlyTabular   bool   `json:"strictly_tabular,omitempty" bson:"strictly_tabular,omitempty"`
	ConsistentFormat  bool   `json:"consistent_format,omitempty" bson:"consistent_format,omitempty"`
	HaveEnrollment    bool   `json:"have_enrollment,omitempty" bson:"have_enrollment,omitempty"`
	ThereIsACapacity  bool   `json:"there_is_a_capacity,omitempty" bson:"there_is_a_capacity,omitempty"`
	HasPosition       bool   `json:"has_position,omitempty" bson:"has_position,omitempty"`
	BaseRevenue       string `json:"base_revenue,omitempty" bson:"base_revenue,omitempty"`
	OtherRecipes      string `json:"other_recipes,omitempty" bson:"other_recipes,omitempty"`
	Expenditure       string `json:"expenditure,omitempty" bson:"expenditure,omitempty"`
}

type Score struct {
	Score             float64 `json:"score" bson:"score"`
	CompletenessScore float64 `json:"completeness_score" bson:"completeness_score"`
	EasinessScore     float64 `json:"easiness_score" bson:"easiness_score"`
}

type AgencyMonthlyInfo struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	AgencyID       string             `json:"aid,omitempty" bson:"aid,omitempty"`
	Month          int                `json:"month,omitempty" bson:"month,omitempty"`
	Year           int                `json:"year,omitempty" bson:"year,omitempty"`
	Meta           *Meta              `json:"meta,omitempty" bson:"meta,omitempy"`
	ExectionTimeMs float64            `json:"exection_time_ms,omitempty" bson:"exection_time_ms,omitempty"`
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	var c config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatal("Error loading config values from .env: ", err.Error())
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(c.MongoURI))
	if err != nil {
		log.Fatal("Error connecting with database: ", err.Error())
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	databases, err := client.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(databases)

	collection := client.Database(c.DBName).Collection(c.MongoMICol)

	res, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal("Error getting result")
	}
	defer res.Close(ctx)
	for res.Next(ctx) {
		var mi AgencyMonthlyInfo
		if err = res.Decode(&mi); err != nil {
			log.Fatal("Error getting mi", err)
		}
		var score = Score{
			Score:             calcScore(*mi.Meta),
			CompletenessScore: calcCompletenessScore(*mi.Meta),
			EasinessScore:     calcEasinessScore(*mi.Meta),
		}
		filter := bson.M{"_id": mi.ID}
		update := bson.M{"$set": bson.M{"score": score}}
		up, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal("Error updating data", err)
		}
		fmt.Printf("Updated %v documents\n", up.ModifiedCount)
	}
}

func calcCriteria(criteria bool, value float64) float64 {
	if criteria {
		return value
	}
	return 0
}

func calcStringCriteria(criteria string, values map[string]float64) float64 {
	for k := range values {
		if criteria == k {
			return values[k]
		}
	}
	return 0
}

func calcCompletenessScore(meta Meta) float64 {
	var score float64 = 0
	var options = map[string]float64{"SUMARIZADO": 0.5, "DETALHADO": 1}

	score = score + calcCriteria(meta.ThereIsACapacity, 1)
	score = score + calcCriteria(meta.HasPosition, 1)
	score = score + calcCriteria(meta.HasPosition, 1)
	score = score + calcStringCriteria(meta.BaseRevenue, options)
	score = score + calcStringCriteria(meta.OtherRecipes, options)
	score = score + calcStringCriteria(meta.Expenditure, options)

	return score / 6
}

func calcEasinessScore(meta Meta) float64 {
	var score float64 = 0
	var options = map[string]float64{
		"ACESSO_DIRETO":          1,
		"AMIGAVEL_PARA_RASPAGEM": 0.5,
		"RASPAGEM_DIFICULTADA":   0.25}

	score = score + calcCriteria(meta.NoLoginRequired, 1)
	score = score + calcCriteria(meta.NoCaptchaRequired, 1)
	score = score + calcStringCriteria(meta.Access, options)
	score = score + calcCriteria(meta.ConsistentFormat, 1)
	score = score + calcCriteria(meta.StrictlyTabular, 1)

	return score / 5
}

func calcScore(meta Meta) float64 {
	var score = 0.0
	var completeness = calcCompletenessScore(meta)
	var easiness = calcEasinessScore(meta)
	score = (completeness + easiness) / 2

	return score
}
