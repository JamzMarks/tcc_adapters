
cd yolo-service
python -m venv .venv
source .venv/bin/activate   # (Linux/macOS)
# ou:
# .venv\Scripts\activate     # (Windows)

uvicorn app.main:app --reload --port 5000