#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Memory Manager - Manage MCP Memory for Claude Code

Commands:
    stats   - Show memory statistics
    list    - List all entities in memory
    search  - Search for entities by name
    clear   - Clear all memory (with confirmation)
    export  - Export memory to JSON file
    import  - Import memory from JSON file

Usage:
    python memory-manager.py stats
    python memory-manager.py list
    python memory-manager.py search "keyword"
    python memory-manager.py clear
    python memory-manager.py export backup.json
    python memory-manager.py import backup.json
"""

import argparse
import json
import os
import sys
import io
from pathlib import Path
from datetime import datetime

# Fix Windows encoding
if sys.platform == "win32":
    try:
        sys.stdout.reconfigure(encoding='utf-8', errors='replace')
    except (AttributeError, io.UnsupportedOperation):
        pass

# Default memory file location for MCP memory server
DEFAULT_MEMORY_PATHS = [
    Path.home() / ".claude" / "memory.json",
    Path.home() / ".config" / "claude-code" / "memory.json",
    Path.cwd() / ".claude" / "memory.json",
]


def find_memory_file() -> Path:
    """Find the memory file location."""
    # Check environment variable first
    env_path = os.environ.get("MCP_MEMORY_PATH")
    if env_path:
        return Path(env_path)

    # Check default locations
    for path in DEFAULT_MEMORY_PATHS:
        if path.exists():
            return path

    # Return first default as fallback
    return DEFAULT_MEMORY_PATHS[0]


def load_memory(path: Path) -> dict:
    """Load memory from file."""
    if not path.exists():
        return {"entities": [], "relations": []}

    try:
        with open(path, 'r', encoding='utf-8') as f:
            return json.load(f)
    except json.JSONDecodeError:
        print(f"Error: Could not parse {path}")
        return {"entities": [], "relations": []}


def save_memory(path: Path, data: dict) -> bool:
    """Save memory to file."""
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        with open(path, 'w', encoding='utf-8') as f:
            json.dump(data, f, indent=2, ensure_ascii=False)
        return True
    except Exception as e:
        print(f"Error saving memory: {e}")
        return False


def cmd_stats(args):
    """Show memory statistics."""
    path = find_memory_file()
    memory = load_memory(path)

    entities = memory.get("entities", [])
    relations = memory.get("relations", [])

    # Calculate sizes
    file_size = path.stat().st_size if path.exists() else 0

    # Count entity types
    entity_types = {}
    total_observations = 0
    for entity in entities:
        etype = entity.get("entityType", "unknown")
        entity_types[etype] = entity_types.get(etype, 0) + 1
        total_observations += len(entity.get("observations", []))

    # Estimate tokens (rough: 1 token ~= 4 chars)
    json_str = json.dumps(memory)
    estimated_tokens = len(json_str) // 4

    print("=" * 50)
    print("Memory Statistics")
    print("=" * 50)
    print(f"File: {path}")
    print(f"File Size: {file_size:,} bytes ({file_size/1024:.1f} KB)")
    print(f"Estimated Tokens: ~{estimated_tokens:,}")
    print()
    print(f"Entities: {len(entities)}")
    print(f"Relations: {len(relations)}")
    print(f"Total Observations: {total_observations}")
    print()

    if entity_types:
        print("Entity Types:")
        for etype, count in sorted(entity_types.items(), key=lambda x: -x[1]):
            print(f"  - {etype}: {count}")

    # Warning if too large
    if estimated_tokens > 20000:
        print()
        print("WARNING: Memory is large and may cause issues!")
        print("Consider running: python memory-manager.py clear")
        print("Or export and clean: python memory-manager.py export backup.json")


def cmd_list(args):
    """List all entities."""
    path = find_memory_file()
    memory = load_memory(path)
    entities = memory.get("entities", [])

    if not entities:
        print("No entities in memory.")
        return

    print(f"Entities ({len(entities)}):")
    print("-" * 50)

    for entity in entities:
        name = entity.get("name", "unnamed")
        etype = entity.get("entityType", "unknown")
        obs_count = len(entity.get("observations", []))
        print(f"  [{etype}] {name} ({obs_count} observations)")


def cmd_search(args):
    """Search for entities."""
    query = args.query.lower()
    path = find_memory_file()
    memory = load_memory(path)
    entities = memory.get("entities", [])

    matches = []
    for entity in entities:
        name = entity.get("name", "").lower()
        etype = entity.get("entityType", "").lower()
        observations = entity.get("observations", [])

        if query in name or query in etype:
            matches.append(entity)
            continue

        for obs in observations:
            if query in obs.lower():
                matches.append(entity)
                break

    if not matches:
        print(f"No entities matching '{args.query}'")
        return

    print(f"Found {len(matches)} matching entities:")
    print("-" * 50)

    for entity in matches:
        name = entity.get("name", "unnamed")
        etype = entity.get("entityType", "unknown")
        print(f"\n[{etype}] {name}")

        for obs in entity.get("observations", [])[:3]:
            preview = obs[:100] + "..." if len(obs) > 100 else obs
            print(f"  - {preview}")

        obs_count = len(entity.get("observations", []))
        if obs_count > 3:
            print(f"  ... and {obs_count - 3} more observations")


def cmd_clear(args):
    """Clear all memory."""
    path = find_memory_file()

    if not path.exists():
        print("Memory file doesn't exist. Nothing to clear.")
        return

    memory = load_memory(path)
    entity_count = len(memory.get("entities", []))
    relation_count = len(memory.get("relations", []))

    if entity_count == 0 and relation_count == 0:
        print("Memory is already empty.")
        return

    print(f"This will delete {entity_count} entities and {relation_count} relations.")

    if not args.yes:
        confirm = input("Are you sure? (yes/no): ")
        if confirm.lower() not in ["yes", "y"]:
            print("Cancelled.")
            return

    # Create backup first
    backup_path = path.with_suffix(f".backup-{datetime.now().strftime('%Y%m%d-%H%M%S')}.json")
    save_memory(backup_path, memory)
    print(f"Backup saved to: {backup_path}")

    # Clear memory
    empty_memory = {"entities": [], "relations": []}
    if save_memory(path, empty_memory):
        print("Memory cleared successfully.")
    else:
        print("Failed to clear memory.")


def cmd_export(args):
    """Export memory to file."""
    path = find_memory_file()
    memory = load_memory(path)

    output_path = Path(args.output)
    if save_memory(output_path, memory):
        entity_count = len(memory.get("entities", []))
        print(f"Exported {entity_count} entities to {output_path}")
    else:
        print("Failed to export memory.")


def cmd_import(args):
    """Import memory from file."""
    input_path = Path(args.input)

    if not input_path.exists():
        print(f"File not found: {input_path}")
        return

    imported = load_memory(input_path)

    if not imported.get("entities") and not imported.get("relations"):
        print("Import file is empty or invalid.")
        return

    path = find_memory_file()
    current = load_memory(path)

    # Merge or replace
    if args.replace:
        new_memory = imported
    else:
        # Merge entities (avoid duplicates by name)
        existing_names = {e.get("name") for e in current.get("entities", [])}
        new_entities = current.get("entities", [])

        for entity in imported.get("entities", []):
            if entity.get("name") not in existing_names:
                new_entities.append(entity)

        # Merge relations
        new_relations = current.get("relations", []) + imported.get("relations", [])

        new_memory = {"entities": new_entities, "relations": new_relations}

    if save_memory(path, new_memory):
        entity_count = len(new_memory.get("entities", []))
        print(f"Memory now has {entity_count} entities")
    else:
        print("Failed to import memory.")


def main():
    parser = argparse.ArgumentParser(
        description="Manage MCP Memory for Claude Code",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )

    subparsers = parser.add_subparsers(dest="command", help="Commands")

    # stats
    subparsers.add_parser("stats", help="Show memory statistics")

    # list
    subparsers.add_parser("list", help="List all entities")

    # search
    search_parser = subparsers.add_parser("search", help="Search entities")
    search_parser.add_argument("query", help="Search query")

    # clear
    clear_parser = subparsers.add_parser("clear", help="Clear all memory")
    clear_parser.add_argument("-y", "--yes", action="store_true", help="Skip confirmation")

    # export
    export_parser = subparsers.add_parser("export", help="Export memory to file")
    export_parser.add_argument("output", help="Output file path")

    # import
    import_parser = subparsers.add_parser("import", help="Import memory from file")
    import_parser.add_argument("input", help="Input file path")
    import_parser.add_argument("--replace", action="store_true", help="Replace instead of merge")

    args = parser.parse_args()

    if not args.command:
        parser.print_help()
        return

    commands = {
        "stats": cmd_stats,
        "list": cmd_list,
        "search": cmd_search,
        "clear": cmd_clear,
        "export": cmd_export,
        "import": cmd_import,
    }

    commands[args.command](args)


if __name__ == "__main__":
    main()
