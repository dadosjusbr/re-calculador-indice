package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/dadosjusbr/indice"
	"github.com/dadosjusbr/proto/coleta"
	"github.com/dadosjusbr/storage"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type config struct {
	MongoURI   string `envconfig:"MONGODB_URI" required:"true"`
	DBName     string `envconfig:"MONGODB_DBNAME" required:"true"`
	MongoMICol string `envconfig:"MONGODB_MICOL" required:"true"`
	MongoAgCol string `envconfig:"MONGODB_AGCOL" required:"true"`
}

var (
	aid  = flag.String("aid", "", "Órgão")
	year = flag.Int("year", 2021, "Ano")
)

func main() {
	flag.Parse()

	if *aid == "" {
		log.Fatal("Flag aid obrigatória")
	}

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Erro ao carregar arquivo .env.")
	}

	var c config
	if err := envconfig.Process("", &c); err != nil {
		log.Fatal("Erro ao carregar parâmetros do arquivo .env: ", err.Error())
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(c.MongoURI))
	if err != nil {
		log.Fatal("Erro ao se conectar com o banco de dados: ", err.Error())
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database(c.DBName).Collection(c.MongoMICol)

	res, err := collection.Find(ctx, bson.M{"aid": *aid, "year": *year})
	if err != nil {
		log.Fatal("Erro ao consultar informações mensais dos órgãos: ", err.Error())
	}
	defer res.Close(ctx)

	fmt.Printf("Atualizando índice de transparência para %s em %d...\n", *aid, *year)
	for res.Next(ctx) {
		var mi storage.AgencyMonthlyInfo
		if err = res.Decode(&mi); err != nil {
			log.Fatalf("[%s/%d/%d] Erro ao obter metadados: %w", mi.AgencyID, mi.Year, mi.Month, err)
		}
		fmt.Printf("%s: %d/%d... ", mi.AgencyID, mi.Month, mi.Year)
		// Quando não houver o dado ou problema na coleta
		if mi.Meta == nil {
			continue
		}
		// a operação inversa é feita no armazenador
		var score = indice.CalcScore(coleta.Metadados{
			TemMatricula:        mi.Meta.HaveEnrollment,
			TemLotacao:          mi.Meta.ThereIsACapacity,
			TemCargo:            mi.Meta.HasPosition,
			ReceitaBase:         coleta.Metadados_OpcoesDetalhamento(coleta.Metadados_OpcoesDetalhamento_value[mi.Meta.BaseRevenue]),
			OutrasReceitas:      coleta.Metadados_OpcoesDetalhamento(coleta.Metadados_OpcoesDetalhamento_value[mi.Meta.OtherRecipes]),
			Despesas:            coleta.Metadados_OpcoesDetalhamento(coleta.Metadados_OpcoesDetalhamento_value[mi.Meta.Expenditure]),
			NaoRequerLogin:      mi.Meta.NoLoginRequired,
			NaoRequerCaptcha:    mi.Meta.NoCaptchaRequired,
			Acesso:              coleta.Metadados_FormaDeAcesso(coleta.Metadados_FormaDeAcesso_value[mi.Meta.Access]),
			FormatoConsistente:  mi.Meta.ConsistentFormat,
			EstritamenteTabular: mi.Meta.StrictlyTabular,
		})
		filter := bson.M{"aid": mi.AgencyID, "year": mi.Year, "month": mi.Month}
		update := bson.M{"$set": bson.M{"score": storage.Score{
			Score:             score.Score,
			CompletenessScore: score.CompletenessScore,
			EasinessScore:     score.EasinessScore,
		}}}
		up, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			log.Fatal("Erro ao atualizar índice", err)
		}
		fmt.Printf("%v docs\n", up.ModifiedCount)
		fmt.Printf("%f %f %f\n", score.Score, score.CompletenessScore, score.EasinessScore)
		time.Sleep(1 * time.Second)
	}
	fmt.Print("Fim.\n")
}
