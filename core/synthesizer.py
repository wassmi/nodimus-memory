import re
import math
import asyncio
import time
from typing import List, Dict, Any, Optional
from collections import defaultdict

class Scrubber:
    """
    PII Redaction using regex.
    """
    def __init__(self):
        self.home_dir = os.path.expanduser("~")
        # Regex for common API keys (simplified patterns)
        self.api_key_patterns = [
            r"sk-[a-zA-Z0-9]{48}",  # OpenAI-like
            r"xox[baprs]-[a-zA-Z0-9]{10,48}", # Slack-like
            r"gh[pousr]-[a-zA-Z0-9]{36}", # GitHub-like
        ]
        self.compiled_patterns = [re.compile(p) for p in self.api_key_patterns]

    def clean(self, text: str) -> str:
        if not text:
            return ""
        
        # Redact Home Directory
        text = text.replace(self.home_dir, "~")
        
        # Redact API Keys
        for pattern in self.compiled_patterns:
            text = pattern.sub("[REDACTED_SECRET]", text)
            
        # Entropy check for other high-entropy strings could go here
        return text

    @staticmethod
    def calculate_entropy(text: str) -> float:
        if not text:
            return 0.0
        prob = [float(text.count(c)) / len(text) for c in dict.fromkeys(list(text))]
        entropy = - sum([p * math.log(p) / math.log(2.0) for p in prob])
        return entropy

class SaliencyFilter:
    """
    Calculates S(e_t) = -log P(e_t | C_t)
    """
    def __init__(self, threshold: float = 2.0):
        self.threshold = threshold
        self.type_counts = defaultdict(int)
        self.total_events = 0

    def is_salient(self, event: Dict[str, Any]) -> bool:
        event_type = event.get("type", "unknown")
        
        # Always keep non-zero exit codes (if applicable)
        if event.get("exit_code", 0) != 0:
            return True

        # Update counts
        self.type_counts[event_type] += 1
        self.total_events += 1

        # Calculate probability P(e_t)
        p_et = self.type_counts[event_type] / self.total_events
        
        # Calculate Saliency (Self-Information)
        # S(e_t) = -log2(P(e_t))
        # If an event is very common, P is high, S is low.
        if p_et > 0:
            saliency = -math.log2(p_et)
        else:
            saliency = float('inf')

        return saliency >= self.threshold

class CoalescenceEngine:
    """
    Groups events and folds them into Episodes.
    """
    def __init__(self, ledger_queue: asyncio.Queue):
        self.active_buffer: List[Dict[str, Any]] = []
        self.ledger_queue = ledger_queue
        self.last_flush_time = time.time()
        self.flush_interval = 300  # 5 minutes

    async def process_event(self, event: Dict[str, Any]):
        self.active_buffer.append(event)
        
        # Check for time delta or buffer size
        current_time = time.time()
        if len(self.active_buffer) > 1000 or current_time - self.last_flush_time > self.flush_interval:
            await self.flush()

    async def flush(self):
        if not self.active_buffer:
            return

        # Group by project_root (assuming event has this field, else default)
        grouped_events = defaultdict(list)
        for event in self.active_buffer:
            root = event.get("project_root", "default")
            grouped_events[root].append(event)

        for root, events in grouped_events.items():
            # "Fold" events into an Episode
            # In a real system, call a micro-LLM here.
            # For now, template-based summary.
            summary = f"Activity in {root}: {len(events)} events processed."
            
            episode = {
                "id": str(uuid.uuid4()),
                "content": summary,
                "metadata": {"event_count": len(events), "root": root},
                "timestamp": time.time(),
                "weight": 1.0,
                "vector": [0.0] * 1536 # Placeholder for embedding
            }
            
            await self.ledger_queue.put(episode)

        self.active_buffer = []
        self.last_flush_time = time.time()

import uuid
import os
