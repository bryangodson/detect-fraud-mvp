# train_model.py
# Train a tiny XGBoost model on synthetic data so you can run locally.
import joblib
import numpy as np
import pandas as pd
from sklearn.model_selection import train_test_split
from xgboost import XGBClassifier
from sklearn.metrics import roc_auc_score

# Synthetic dataset with features we expect
N = 5000
rng = np.random.RandomState(42)
amount = rng.exponential(scale=200.0, size=N)               # transaction amounts
account_age = rng.randint(0, 365, size=N)                  # account days
is_new_device = rng.binomial(1, 0.1, size=N)               # 10% new devices
country_risk = rng.binomial(1, 0.02, size=N)               # 2% high-risk country

# Simple label: higher chance of fraud when amount large and new device / country risk
logit =  -3 + 0.005*amount + 1.2*is_new_device + 2.5*country_risk - 0.002*account_age
prob = 1 / (1 + np.exp(-logit))
y = rng.binomial(1, prob)

df = pd.DataFrame({
    "amount": amount,
    "account_age": account_age,
    "is_new_device": is_new_device,
    "country_risk": country_risk,
    "label": y
})

X = df[["amount", "account_age", "is_new_device", "country_risk"]]
y = df["label"]

X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

model = XGBClassifier(use_label_encoder=False, eval_metric="logloss", n_estimators=100, max_depth=4)
model.fit(X_train, y_train)

preds = model.predict_proba(X_test)[:,1]
print("AUC:", roc_auc_score(y_test, preds))

# Save model and optionally save scaler or metadata
joblib.dump(model, "model.joblib")
print("Model saved to model.joblib")
