#!/usr/bin/env python3
"""
Sync Warframe JSON data to MongoDB.

This script reads JSON files from the ./json directory and syncs them to MongoDB.
It handles:
- Inserts: New items are added
- Updates: Existing items are updated based on uniqueName
- Deletes: Items in MongoDB that no longer exist in JSON are removed
"""

import json
import os
import sys
from pathlib import Path
from typing import Any

from pymongo import MongoClient, UpdateOne
from pymongo.collection import Collection
from pymongo.database import Database

# Files to skip (large aggregated/translation files)
SKIP_FILES = {"All.json", "i18n.json"}

# Default MongoDB connection settings
DEFAULT_MONGO_URI = "mongodb://localhost:27017"
DEFAULT_DATABASE = "warframe"


def load_json_file(file_path: Path) -> list[dict[str, Any]]:
    """Load and parse a JSON file."""
    with open(file_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    if not isinstance(data, list):
        raise ValueError(f"Expected array in {file_path}, got {type(data).__name__}")

    return data


def get_collection_name(file_path: Path) -> str:
    """Convert filename to collection name (lowercase, no extension)."""
    return file_path.stem.lower().replace("-", "_")


def sync_collection(
    collection: Collection,
    items: list[dict[str, Any]],
    dry_run: bool = False
) -> dict[str, int]:
    """
    Sync items to a MongoDB collection.

    Uses uniqueName as the unique identifier for each document.
    Returns statistics about the sync operation.
    """
    stats = {"inserted": 0, "updated": 0, "deleted": 0, "unchanged": 0}

    # Build a set of uniqueNames from the JSON data
    json_unique_names = set()
    bulk_operations = []

    for item in items:
        unique_name = item.get("uniqueName")
        if not unique_name:
            # Skip items without uniqueName
            continue

        json_unique_names.add(unique_name)

        # Prepare upsert operation
        bulk_operations.append(
            UpdateOne(
                {"uniqueName": unique_name},
                {"$set": item},
                upsert=True
            )
        )

    if dry_run:
        # Count what would happen
        existing_docs = {
            doc["uniqueName"]
            for doc in collection.find({}, {"uniqueName": 1})
            if "uniqueName" in doc
        }

        new_items = json_unique_names - existing_docs
        to_delete = existing_docs - json_unique_names
        to_update = existing_docs & json_unique_names

        stats["inserted"] = len(new_items)
        stats["updated"] = len(to_update)
        stats["deleted"] = len(to_delete)
        return stats

    # Execute bulk upserts
    if bulk_operations:
        result = collection.bulk_write(bulk_operations, ordered=False)
        stats["inserted"] = result.upserted_count
        stats["updated"] = result.modified_count
        stats["unchanged"] = result.matched_count - result.modified_count

    # Delete items no longer in JSON
    existing_unique_names = {
        doc["uniqueName"]
        for doc in collection.find({}, {"uniqueName": 1})
        if "uniqueName" in doc
    }

    to_delete = existing_unique_names - json_unique_names
    if to_delete:
        delete_result = collection.delete_many({"uniqueName": {"$in": list(to_delete)}})
        stats["deleted"] = delete_result.deleted_count

    return stats


def sync_all(
    json_dir: Path,
    mongo_uri: str,
    database_name: str,
    dry_run: bool = False
) -> dict[str, dict[str, int]]:
    """
    Sync all JSON files to MongoDB.

    Returns statistics for each collection.
    """
    client = MongoClient(mongo_uri)
    db: Database = client[database_name]

    all_stats = {}

    # Get all JSON files
    json_files = sorted(json_dir.glob("*.json"))

    for json_file in json_files:
        if json_file.name in SKIP_FILES:
            print(f"Skipping {json_file.name}")
            continue

        collection_name = get_collection_name(json_file)
        print(f"Processing {json_file.name} -> {collection_name}...", end=" ")

        try:
            items = load_json_file(json_file)
            collection = db[collection_name]

            # Create index on uniqueName if it doesn't exist
            if not dry_run:
                collection.create_index("uniqueName", unique=True, sparse=True)

            stats = sync_collection(collection, items, dry_run=dry_run)
            all_stats[collection_name] = stats

            print(
                f"inserted={stats['inserted']}, "
                f"updated={stats['updated']}, "
                f"deleted={stats['deleted']}, "
                f"unchanged={stats['unchanged']}"
            )
        except Exception as e:
            print(f"ERROR: {e}")
            all_stats[collection_name] = {"error": str(e)}

    client.close()
    return all_stats


def print_summary(stats: dict[str, dict[str, int]]) -> None:
    """Print a summary of all sync operations."""
    total_inserted = 0
    total_updated = 0
    total_deleted = 0
    total_unchanged = 0
    errors = 0

    for collection_stats in stats.values():
        if "error" in collection_stats:
            errors += 1
        else:
            total_inserted += collection_stats.get("inserted", 0)
            total_updated += collection_stats.get("updated", 0)
            total_deleted += collection_stats.get("deleted", 0)
            total_unchanged += collection_stats.get("unchanged", 0)

    print("\n" + "=" * 50)
    print("SUMMARY")
    print("=" * 50)
    print(f"Collections processed: {len(stats)}")
    print(f"Total inserted: {total_inserted}")
    print(f"Total updated: {total_updated}")
    print(f"Total deleted: {total_deleted}")
    print(f"Total unchanged: {total_unchanged}")
    if errors:
        print(f"Errors: {errors}")


def main() -> int:
    """Main entry point."""
    import argparse

    parser = argparse.ArgumentParser(
        description="Sync Warframe JSON data to MongoDB"
    )
    parser.add_argument(
        "--json-dir",
        type=Path,
        default=Path(__file__).parent / "json",
        help="Path to JSON directory (default: ./json)"
    )
    parser.add_argument(
        "--mongo-uri",
        default=os.environ.get("MONGO_URI", DEFAULT_MONGO_URI),
        help=f"MongoDB connection URI (default: {DEFAULT_MONGO_URI})"
    )
    parser.add_argument(
        "--database",
        default=os.environ.get("MONGO_DATABASE", DEFAULT_DATABASE),
        help=f"MongoDB database name (default: {DEFAULT_DATABASE})"
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without making changes"
    )

    args = parser.parse_args()

    if not args.json_dir.exists():
        print(f"Error: JSON directory not found: {args.json_dir}", file=sys.stderr)
        return 1

    print(f"JSON directory: {args.json_dir}")
    print(f"MongoDB URI: {args.mongo_uri}")
    print(f"Database: {args.database}")
    if args.dry_run:
        print("DRY RUN MODE - No changes will be made")
    print()

    stats = sync_all(
        json_dir=args.json_dir,
        mongo_uri=args.mongo_uri,
        database_name=args.database,
        dry_run=args.dry_run
    )

    print_summary(stats)

    return 0


if __name__ == "__main__":
    sys.exit(main())
