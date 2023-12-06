package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/cristosal/pay"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	whsec  = os.Getenv("STRIPE_WEBHOOK_SECRET")
	apikey = os.Getenv("STRIPE_API_KEY")
)

func main() {
	db, err := sql.Open("pgx", os.Getenv("CONNECTION_STRING"))
	if err != nil {
		log.Fatal(err)
	}

	provider := pay.NewStripeProvider(&pay.StripeConfig{
		Repo:          pay.NewEntityRepo(db),
		Key:           apikey,
		WebhookSecret: whsec,
	})

	if err := provider.Init(context.Background()); err != nil {
		log.Fatal(err)
	}

	fmt.Print("syncing...")
	if err := provider.Sync(); err != nil {
		log.Fatal(err)
	}
	fmt.Print(" [OK]\n")
	fmt.Println("listening on port 8080")
	http.HandleFunc("/", handleHome())
	http.HandleFunc("/webhook", provider.Webhook())
	http.HandleFunc("/plans", handlePlans(provider))
	http.HandleFunc("/plans/new", handlePlansNew(provider))
	http.HandleFunc("/plans/delete", handlePlansDelete(provider))
	http.HandleFunc("/prices/", handlePrices(provider))
	http.HandleFunc("/prices/new", handlePricesNew(provider))
	http.HandleFunc("/sync", handleSync(provider))
	http.ListenAndServe("127.0.0.1:8080", nil)
}

func handleSync(p *pay.StripeProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := p.Sync(); err != nil {
			writeErr(w, err)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func handleHome() http.HandlerFunc {
	html := `
<html lang="en">
	<head>
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css">
		<title>Pay</title>
	</head>
	<body>
		<main class="container">
			<h1>Pay</h1>
			<p>Payment Microservice</p>
			<nav>
				<ol>
					<li><a href="/plans">Plans</a></li>
					<li><a href="/prices">Prices</a></li>
					<li><a href="/customers">Customers</a></li>
					<li><a href="/subscriptions">Subscriptions</a></li>
					<li><a href="/events">Webhook Events</a></li>
					<li><a href="/sync">Sync</a></li>
				</ol>
			</nav>
		</main>
	</body>
</html>`

	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}
}

func handlePlansNew(p *pay.StripeProvider) http.HandlerFunc {
	html := `
<html lang="en">
	<head>
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css">
		<title>New Plan</title>
	</head>
	<body>
		<main class="container">
			<form method="POST" action="/plans">
				<h2>Add Plan</h2>
				<div>
					<label for="name">Name</label>
					<input id="name" name="name" type="text">
				</div>
				<div>
					<label for="description">Description</label>
					<input id="description" name="description" type="text">
				</div>
				<div>
					<label for="active">
						Active
						<input id="active" name="active" type="checkbox">
					</label>
				</div>
				<br>
				<input type="submit" value="Create Plan">
			</form>
		</main>
	</body>
</html>`

	t := template.Must(template.New("").Parse(html))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				log.Printf("error while parsing form: %v", err)
				http.Redirect(w, r, "/plans", http.StatusSeeOther)
				return
			}

			name := r.Form.Get("name")
			desc := r.Form.Get("description")
			active := r.Form.Get("active") == "on"

			err := p.AddPlan(&pay.Plan{
				Name:        name,
				Description: desc,
				Active:      active,
			})

			if err != nil {
				log.Printf("error while adding plan: %v", err)
			}

			http.Redirect(w, r, "/plans", http.StatusSeeOther)
			return
		}

		t.Execute(w, nil)
	}
}

func handlePlans(p *pay.StripeProvider) http.HandlerFunc {
	html := `
<html lang="en">
	<head>
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css">
		<title>Plans</title>
	</head>
	<body>
		<main class="container">
			<h1>Plans</h1>
			<a href="/plans/new">Add Plan</a>
			<br>
			<table>
				<thead>
					<th>Name</th>
					<th>Description</th>
					<th>Provider</th>
					<th>ProviderID</th>
					<th>Actions</th>
				</thead>
				<tbody>
				{{- range .Plans -}}
				<tr>
					<td>{{ .Name }}</td>
					<td>{{ .Description }}</td>
					<td>{{ .Provider }}</td>
					<td>{{ .ProviderID }}</td>
					<td><a href="/plans/delete?id={{ .ID }}">Del</a></td>
				</tr>
				{{- end -}}
				</tbody>
			</table>
		</main>
	</body>
</html>`

	t := template.Must(template.New("").Parse(html))

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			if err := r.ParseForm(); err != nil {
				log.Printf("error while parsing form: %v", err)
				http.Redirect(w, r, "/plans", http.StatusSeeOther)
				return
			}

			name := r.Form.Get("name")
			desc := r.Form.Get("description")
			active := r.Form.Get("active") == "on"

			err := p.AddPlan(&pay.Plan{
				Name:        name,
				Description: desc,
				Active:      active,
			})

			if err != nil {
				log.Printf("error while adding plan: %v", err)
			}

			http.Redirect(w, r, "/plans", http.StatusSeeOther)
			return
		}

		plans, err := p.Repo().ListPlans()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		err = t.Execute(w, map[string]any{
			"Plans": plans,
		})
		if err != nil {
			w.Write([]byte(err.Error()))
		}
	}
}

func handlePlansDelete(p *pay.StripeProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idQuery := r.URL.Query().Get("id")
		id, err := strconv.ParseInt(idQuery, 10, 64)
		if err != nil {
			writeErr(w, err)
			return
		}

		pl, err := p.Repo().GetPlanByID(id)
		if err != nil {
			writeErr(w, err)
			return
		}

		if err := p.RemovePlan(pl); err != nil {
			writeErr(w, err)
			return
		}

		http.Redirect(w, r, "/plans", http.StatusSeeOther)
	}
}

func handlePrices(p *pay.StripeProvider) http.HandlerFunc {
	html := `
<html lang="en">
	<head>
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css">
		<title>Prices</title>
	</head>
	<body>
		<main class="container">
			<h1>Prices</h1>
			<a href="/prices/new">Add</a>
			<br>
			<table>
				<thead>
					<th>ProviderID</th>
					<th>Currency</th>
					<th>Amount</th>
					<th>Schedule</th>
					<th>PlanID</th>
					<th>Trial Days</th>
				</thead>
				<tbody>
					{{ range .Prices }}
					<tr>
						<td>{{ .ProviderID }}</td>
						<td>{{ .Currency }}</td>
						<td>{{ .Amount }}</td>
						<td>{{ .Schedule }}</td>
						<td>{{ .PlanID }}</td>
						<td>{{ .TrialDays }}</td>
					</tr>
					{{ end }}
				</tbody>
			</table>
		</main>
	</body>
</html>`

	t := template.Must(template.New("prices").Parse(html))

	return func(w http.ResponseWriter, r *http.Request) {
		plans, err := p.Repo().ListPlans()
		if err != nil {
			writeErr(w, err)
			return
		}

		prices, err := p.Repo().ListPrices()
		if err != nil {
			writeErr(w, err)
			return
		}

		err = t.Execute(w, map[string]any{
			"Plans":   plans,
			"Prices":  prices,
			"Monthly": pay.PricingMonthly,
			"Annual":  pay.PricingAnnual,
		})

		if err != nil {
			writeErr(w, err)
			return
		}
	}
}

func handlePricesNew(p *pay.StripeProvider) http.HandlerFunc {
	t := createTemplate(`
<!DOCTYPE html>
<html lang="en">
	<head>
		<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/@picocss/pico@1/css/pico.min.css">
		<title>New Price</title>
	</head>
	<body>
		<main class="container">
			<form method="post" action="/prices/new">
				<h2>Add Price</h2>
				<div>
					<label for="currency">Currency</label>
					<input id="currency" name="currency" type="text">
				</div>
				<div>
					<label for="amount">Amount</label>
					<input id="amount" name="amount" type="number">
				</div>
				<div>
					<label for="trial_days">Trial Days</label>
					<input id="trial_days" name="trial_days" type="number">
				</div>
				<div>
					<label for="plan_id">Plan</label>
					<select id="plan_id" name="plan_id">
					{{- range .Plans -}}
						<option value="{{ .ID }}">{{ .Name }}</option>
					{{- end -}}
					</select>
				</div>
				<div>
					<label for="schedule">Schedule</label>
					<select id="schedule" name="schedule">
						<option value="{{ .Monthly }}">Monthly</option>
						<option value="{{ .Annual }}">Annual</option>
					</select>
				</div>
				<br>
				<input type="submit" value="Create Price">
			</form>
		</main>
	</body>
</html>`)

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				writeErr(w, err)
				return
			}

			var (
				planID    = r.FormValue("plan_id")
				amount    = r.FormValue("amount")
				currency  = r.FormValue("currency")
				schedule  = r.FormValue("schedule")
				trialDays = r.FormValue("trial_days")
			)

			parsedAmount, err := strconv.Atoi(amount)
			if err != nil {
				writeErr(w, fmt.Errorf("error parsing amount: %w", err))
				return
			}

			parsedPlanID, err := strconv.Atoi(planID)
			if err != nil {
				writeErr(w, fmt.Errorf("error parsing plan id: %w", err))
				return
			}

			parsedTrialDays, err := strconv.Atoi(trialDays)
			if err != nil {
				writeErr(w, fmt.Errorf("error parsing trial days: %w", err))
				return
			}

			pr := pay.Price{
				Amount:    int64(parsedAmount),
				PlanID:    int64(parsedPlanID),
				Currency:  currency,
				Schedule:  schedule,
				TrialDays: parsedTrialDays,
			}

			if err := p.AddPrice(&pr); err != nil {
				writeErr(w, fmt.Errorf("error adding price: %w", err))
				return
			}

			http.Redirect(w, r, "/prices", http.StatusSeeOther)
		default:
			plans, err := p.Repo().ListPlans()
			if err != nil {
				writeErr(w, err)
				return
			}

			renderTemplate(t, w, map[string]any{
				"Plans":   plans,
				"Monthly": pay.PricingMonthly,
				"Annual":  pay.PricingAnnual,
			})
		}
	}
}

func handleCustomers(p *pay.StripeProvider) http.HandlerFunc {
	t := createTemplate(`
<!DOCTYPE html>
<html lang="en">
	<head>
		<title>Customers</title>
	</head>
	<body>
		<h1>Customers</h1>
		<table>
		</table>
	</body>
</html>
	`)

	return handle(func(w http.ResponseWriter, r *http.Request) error {
		return t.Execute(w, nil)
	})
}

func writeErr(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
}

func createTemplate(html string) *template.Template {
	return template.Must(template.New("").Parse(html))
}

type wrappedHandlerFunc func(w http.ResponseWriter, r *http.Request) error

func handle(h wrappedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			writeErr(w, err)
		}
	}

}

func renderTemplate(t *template.Template, w http.ResponseWriter, data map[string]any) {
	if err := t.Execute(w, data); err != nil {
		writeErr(w, err)
	}
}
