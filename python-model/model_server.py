# model_server.py
from fastapi import FastAPI
from pydantic import BaseModel
import joblib
import pandas as pd

app = FastAPI()

# Load model on start
model = joblib.load("model.joblib")

class ScoreRequest(BaseModel):
    # fields we accept from the Go service
    amount: float
    account_age: int
    is_new_device: bool
    country: str = None  # optional; we map to country_risk inside

@app.post("/score")
def score(req: ScoreRequest):
    # map country to simple country_risk for demo
    high_risk_countries = {"RU", "KP", "NG"}
    country_risk = 1 if (req.country and req.country.upper() in high_risk_countries) else 0

    # Build DataFrame for model
    df = pd.DataFrame([{
        "amount": req.amount,
        "account_age": req.account_age,
        "is_new_device": 1 if req.is_new_device else 0,
        "country_risk": country_risk
    }])

    prob = model.predict_proba(df)[:,1][0]
    return {"score": float(prob)}
