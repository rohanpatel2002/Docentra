import sys
import json
import argparse
import logging
from typing import List
from fastembed import TextEmbedding

logging.basicConfig(level=logging.ERROR, stream=sys.stderr)
logger = logging.getLogger(__name__)
def simple_chunk_text(text: str, chunk_size: int = 1000, overlap: int = 200) -> List[str]:
    if not text:
        return []
    chunks = []
    start = 0
    while start < len(text):
        end = start + chunk_size
        chunks.append(text[start:end])
        start += chunk_size - overlap
    return chunks
def process_text(text: str):
    try:
        # Load model 
        model = TextEmbedding(model_name="BAAI/bge-small-en-v1.5")
        # Chunking
        text_chunks = simple_chunk_text(text)
        # Embedding
        embeddings_gen = model.embed(text_chunks)
        embeddings = [emb.tolist() for emb in embeddings_gen]
        # Format result
        result = {
            "chunks": [
                {"content": chunk, "embedding": emb}
                for chunk, emb in zip(text_chunks, embeddings)
            ]
        }
        
        print(json.dumps(result))
        
    except Exception as e:
        logger.error(f"Execution Error: {str(e)}")
        sys.exit(1)

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="AI Document Assistant - Vector CLI")
    parser.add_argument("--text", type=str, required=True, help="Text to chunk and embed")
    args = parser.parse_args()
    process_text(args.text)