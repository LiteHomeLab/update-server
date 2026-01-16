#!/usr/bin/env python3
"""
Pushover notification hook for Claude Code.

Sends notifications when:
- Task completes (Stop hook)
- Attention needed (Notification hook for permission/idle prompts)
"""

import json
import os
import subprocess
import sys
from datetime import datetime
from pathlib import Path


# Setup logging
def get_log_path() -> Path:
    """Get the debug log file path."""
    # Use the script's directory for logs
    script_dir = Path(__file__).parent
    return script_dir / "debug.log"


def log(message: str) -> None:
    """Write a message to the debug log with timestamp."""
    try:
        log_path = get_log_path()
        timestamp = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        with open(log_path, "a", encoding="utf-8") as f:
            f.write(f"[{timestamp}] {message}\n")
    except Exception:
        pass


def send_pushover(title: str, message: str, priority: int = 0) -> bool:
    """
    Send a notification via Pushover API using curl.

    Args:
        title: Notification title
        message: Notification message body
        priority: Message priority (-2 to 2, default 0)

    Returns:
        True if successful, False otherwise
    """
    log(f"send_pushover called: title='{title}', priority={priority}")

    token = os.environ.get("PUSHOVER_TOKEN")
    user = os.environ.get("PUSHOVER_USER")

    if not token or not user:
        log(f"ERROR: Missing env vars - TOKEN={bool(token)}, USER={bool(user)}")
        return False

    log(f"Environment variables found - TOKEN: {token[:10]}..., USER: {user[:10]}...")

    # Create a temp file to capture response body for debugging
    import tempfile
    with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.txt') as f:
        response_file = f.name

    try:
        # Convert literal \n to actual newlines for the message
        message = message.replace("\\n", "\n")

        # Build curl command with URL encoding
        cmd = [
            "curl",
            "-s",
            "-o", response_file,
            "-w", "%{http_code}",
            "https://api.pushover.net/1/messages.json",
            "--data-urlencode", f"token={token}",
            "--data-urlencode", f"user={user}",
            "--data-urlencode", f"title={title}",
            "--data-urlencode", f"message={message}",
            "-d", f"priority={priority}",
        ]

        log(f"Executing curl command...")
        log(f"Command: {' '.join(cmd[:4])} ...")  # Log first part of command

        result = subprocess.run(cmd, capture_output=True, text=True, timeout=10)

        log(f"Curl return code: {result.returncode}")
        http_code = result.stdout.strip()
        log(f"HTTP Status Code: {http_code}")

        # Read and log response body
        try:
            with open(response_file, 'r', encoding='utf-8') as f:
                response_body = f.read()
            log(f"API Response: {response_body}")

            # Parse JSON for detailed error info
            try:
                response_json = json.loads(response_body)
                if response_json.get("status") == 1:
                    log(f"Request successful - ID: {response_json.get('request', 'N/A')}")
                    return True
                else:
                    log("ERROR: API returned status != 1")
                    if "errors" in response_json:
                        for error in response_json["errors"]:
                            log(f"API Error: {error}")
                    return False
            except json.JSONDecodeError:
                log("WARNING: Could not parse response as JSON")

        except Exception as e:
            log(f"WARNING: Could not read response file: {e}")

        # Fallback: check HTTP code
        success = result.returncode == 0 and http_code == "200"
        if not success:
            if http_code == "400":
                log("ERROR: HTTP 400 - Bad Request (check token/user key)")
            elif http_code == "401":
                log("ERROR: HTTP 401 - Unauthorized (invalid token)")
            elif http_code == "404":
                log("ERROR: HTTP 404 - User not found")
        log(f"Request successful: {success}")
        return success

    except subprocess.TimeoutExpired:
        log("ERROR: Curl request timed out")
        return False
    except FileNotFoundError:
        log("ERROR: Curl not found")
        return False
    except Exception as e:
        log(f"ERROR: Exception in send_pushover: {e}")
        return False
    finally:
        # Clean up temp file
        try:
            Path(response_file).unlink(missing_ok=True)
        except Exception:
            pass


def get_project_name(cwd: str) -> str:
    """
    Extract project name from working directory path.

    Args:
        cwd: Current working directory

    Returns:
        Project name or fallback string
    """
    try:
        name = os.path.basename(os.path.normpath(cwd))
        log(f"Extracted project name: {name} from {cwd}")
        return name
    except Exception as e:
        log(f"ERROR getting project name: {e}")
        return "Unknown Project"


def summarize_conversation(session_id: str, cwd: str) -> str:
    """
    Generate a summary of the conversation using Claude CLI.

    Args:
        session_id: The session identifier
        cwd: Current working directory

    Returns:
        Summary string or fallback message
    """
    log(f"summarize_conversation called for session {session_id}")

    cache_dir = Path(cwd) / ".claude" / "cache"
    cache_file = cache_dir / f"session-{session_id}.jsonl"

    # Fallback: extract last user message
    fallback_summary = "Task completed"

    if not cache_file.exists():
        log(f"Cache file not found: {cache_file}")
        return fallback_summary

    try:
        lines = cache_file.read_text(encoding="utf-8").strip().split("\n")
        log(f"Cache file has {len(lines)} lines")

        if not lines or lines == [""]:
            log("Cache file is empty")
            return fallback_summary

        # Get last user message as fallback
        for line in reversed(lines):
            try:
                data = json.loads(line)
                if data.get("type") == "user_prompt_submit":
                    content = data.get("prompt", "")
                    if content:
                        # Truncate to reasonable length
                        fallback_summary = (
                            content[:100] + "..." if len(content) > 100 else content
                        )
                        log(f"Using fallback summary from user message")
                        break
            except json.JSONDecodeError:
                continue

        # Try to use Claude CLI for summarization
        try:
            conversation_text = "\n".join(lines)
            prompt = f"""Summarize this conversation in one concise sentence (max 15 words):

{conversation_text}

Summary:"""

            log("Attempting Claude CLI summarization...")
            result = subprocess.run(
                ["claude", "-p", prompt],
                capture_output=True,
                text=True,
                timeout=30,
                cwd=cwd,
            )

            if result.returncode == 0 and result.stdout.strip():
                summary = result.stdout.strip()
                if len(summary) < 200:
                    log(f"Claude CLI summary: {summary}")
                    return summary
                else:
                    log(f"Claude CLI summary too long ({len(summary)} chars), using fallback")

            log(f"Claude CLI failed - return code: {result.returncode}")

        except subprocess.TimeoutExpired:
            log("Claude CLI timed out")
        except FileNotFoundError:
            log("Claude CLI not found")
        except Exception as e:
            log(f"Claude CLI exception: {e}")

        return fallback_summary

    except Exception as e:
        log(f"ERROR in summarize_conversation: {e}")
        return fallback_summary


def main() -> None:
    """Main hook handler."""
    log("=" * 60)

    # Force UTF-8 encoding for stdin on all platforms (Windows encoding fix)
    if hasattr(sys.stdin, 'reconfigure'):
        sys.stdin.reconfigure(encoding='utf-8')
        log(f"Stdin encoding configured: {sys.stdin.encoding}")
    else:
        log("WARNING: stdin.reconfigure not available (Python < 3.7)")

    log(f"Hook script started - Event: Processing")

    # Read hook event from stdin
    try:
        stdin_data = sys.stdin.read()
        log(f"Stdin read successfully, length: {len(stdin_data)}")
    except Exception as e:
        log(f"ERROR reading stdin: {e}")
        return

    if not stdin_data:
        log("ERROR: stdin is empty")
        return

    log(f"Stdin content: {stdin_data[:200]}...")

    # Fix Windows paths in JSON (backslashes need to be escaped)
    stdin_data = stdin_data.replace("\\", "\\\\")

    try:
        hook_input = json.loads(stdin_data)
        log(f"JSON parsed successfully")
    except json.JSONDecodeError as e:
        log(f"ERROR: JSON decode failed: {e}")
        return

    hook_event = hook_input.get("hook_event_name", "")
    session_id = hook_input.get("session_id", "")
    cwd = hook_input.get("cwd", os.getcwd())

    log(f"Event: {hook_event}, Session: {session_id}, CWD: {cwd}")

    if not session_id:
        log("ERROR: No session_id in input")
        return

    if hook_event == "UserPromptSubmit":
        log("Processing UserPromptSubmit event")
        # Record user input to cache
        cache_dir = Path(cwd) / ".claude" / "cache"
        cache_dir.mkdir(parents=True, exist_ok=True)

        cache_file = cache_dir / f"session-{session_id}.jsonl"

        try:
            entry = {
                "type": "user_prompt_submit",
                "prompt": hook_input.get("prompt", ""),
                "timestamp": hook_input.get("timestamp", ""),
            }

            with open(cache_file, "a", encoding="utf-8") as f:
                f.write(json.dumps(entry) + "\n")
            log(f"User prompt cached to {cache_file}")
        except (OSError, IOError) as e:
            log(f"ERROR caching user prompt: {e}")

    elif hook_event == "Stop":
        log("Processing Stop event")
        # Send task completion notification
        project_name = get_project_name(cwd)
        summary = summarize_conversation(session_id, cwd)

        title = f"[{project_name}] Task Complete"
        message = f"Session: {session_id}\\nSummary: {summary}"

        log(f"Sending notification: {title}")
        send_pushover(title, message, priority=0)

        log(f"Message stats: chars={len(message)}, bytes={len(message.encode('utf-8'))}")

        # Clean up cache
        cache_file = Path(cwd) / ".claude" / "cache" / f"session-{session_id}.jsonl"
        try:
            cache_file.unlink(missing_ok=True)
            log(f"Cache file cleaned up: {cache_file}")
        except OSError as e:
            log(f"ERROR cleaning up cache: {e}")

    elif hook_event == "Notification":
        log("Processing Notification event")
        # Log full input for debugging
        log(f"Full Notification input: {json.dumps(hook_input, ensure_ascii=False)}")
        # Get notification type (correct field name from docs)
        notification_type = hook_input.get("notification_type", "notification")
        log(f"Notification type: {notification_type}")

        # Skip idle_prompt notifications (CLI idle for 60+ seconds)
        if notification_type == "idle_prompt":
            log("Skipping idle_prompt notification - not sending pushover")
            return

        # Get notification message (correct field name from docs)
        notification_message = hook_input.get("message", "")

        project_name = get_project_name(cwd)

        title = f"[{project_name}] Attention Needed"

        # Build message from notification
        details = notification_message if notification_message else "No additional details provided"

        message = f"Session: {session_id}\\nType: {notification_type}\\n{details}"

        log(f"Sending attention notification: {title}")
        # Higher priority for attention needed
        send_pushover(title, message, priority=1)

        log(f"Message stats: chars={len(message)}, bytes={len(message.encode('utf-8'))}")
    else:
        log(f"WARNING: Unknown hook event type: {hook_event}")

    log(f"Hook script completed")
    log("=" * 60)


if __name__ == "__main__":
    main()
