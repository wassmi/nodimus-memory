import asyncio
import typer
import os
from typing import Dict, Any

from core.observer import FSObserver, TerminalObserver
from core.synthesizer import CoalescenceEngine, Scrubber, SaliencyFilter
from core.ledger import Ledger
from agent_server.server import set_ledger, mcp

app = typer.Typer()

async def pipeline_worker(
    input_queue: asyncio.Queue,
    synthesizer: CoalescenceEngine,
    scrubber: Scrubber,
    saliency: SaliencyFilter
):
    """
    Reads raw events, scrubs, filters, and passes to coalescence engine.
    """
    print("Pipeline Worker started")
    while True:
        event = await input_queue.get()
        try:
            # 1. Scrub PII (if event has content/path)
            if "path" in event:
                event["path"] = scrubber.clean(event["path"])
            if "content" in event:
                event["content"] = scrubber.clean(event["content"])

            # 2. Saliency Filter
            if saliency.is_salient(event):
                # 3. Coalesce
                await synthesizer.process_event(event)
        except Exception as e:
            print(f"Pipeline Error: {e}")
        finally:
            input_queue.task_done()

async def ledger_worker(queue: asyncio.Queue, ledger: Ledger):
    """
    Reads episodes from synthesizer and writes to Ledger.
    """
    print("Ledger Worker started")
    while True:
        episode = await queue.get()
        try:
            await ledger.add_episode(episode)
            print(f"Persisted Episode: {episode['id']}")
        except Exception as e:
            print(f"Ledger Worker Error: {e}")
        finally:
            queue.task_done()

async def main_loop():
    print("Starting NODIMUS Cognitive Memory...")
    
    # 1. Initialize Queues
    raw_event_queue = asyncio.Queue()
    ledger_queue = asyncio.Queue()

    # 2. Initialize Components
    ledger = Ledger()
    set_ledger(ledger) # Inject into MCP

    scrubber = Scrubber()
    saliency = SaliencyFilter()
    synthesizer = CoalescenceEngine(ledger_queue)

    fs_observer = FSObserver(raw_event_queue)
    term_observer = TerminalObserver(raw_event_queue)

    # 3. Start Subsystems
    try:
        await asyncio.gather(
            fs_observer.start(),
            term_observer.start(),
            ledger.decay_loop(),
            pipeline_worker(raw_event_queue, synthesizer, scrubber, saliency),
            ledger_worker(ledger_queue, ledger),
            mcp.run_stdio_async() # Starts the MCP server (stdio)
        )
    except asyncio.CancelledError:
        print("Shutting down...")
        await fs_observer.stop()
        await term_observer.stop()

@app.command()
def start():
    """
    Start the NODIMUS daemon.
    """
    asyncio.run(main_loop())

if __name__ == "__main__":
    app()
