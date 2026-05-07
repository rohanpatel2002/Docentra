from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
import logging

app = FastAPI(title="AI Document Assistant - Embedding Service", version="1.0.0")

# Generic logging securely
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)
# Context model wrapper
class EmbedRequest(BaseModel):
    text: str
class EmbedResponse(BaseModel):
    embedding: list[float]  
# Global embedding model
model = None
@app.on_event("startup")
async def load_model():
    global model
    logger.info("Loading sentence transformer model securely into memory...")
    model = SentenceTransformer('all-MiniLM-L6-v2') 
    logger.info("Model securely loaded and ready")
@app.post("/embed", response_model=EmbedResponse)
async def embed_text(req: EmbedRequest):
    if model is None:
        raise HTTPException(status_code=503, detail="Model not initialized")
    if not req.text or len(req.text.strip()) == 0:
        raise HTTPException(status_code=400, detail="Empty text provided for embedding")
    try:
        embedding_data = model.encode(req.text).tolist()
        return EmbedResponse(embedding=embedding_data)
    except Exception as e:
        logger.error(f"Failed to generate embedding: {str(e)}")
        raise HTTPException(status_code=500, detail="Internal embedding server error")
@app.get("/health")
async def health_check():
    return {"status": "healthy", "model_loaded": model is not None}
