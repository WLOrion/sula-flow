import os
import json
import csv
import sys

def load_countries_map(countries_file: str) -> dict:
    country_map = {}
    with open(countries_file, "r", encoding="utf-8") as f:
        reader = csv.DictReader(f)
        for row in reader:
            country_map[row["name"]] = row["id"]
    return country_map


def save_csv(path: str, header: list[str], rows: list) -> None:
    with open(path, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=header)
        writer.writeheader()
        writer.writerows(rows)


def save_json(path: str, data: list[object]) -> None:
    with open(path, "w", encoding="utf-8") as f:
        json.dump(data, f, ensure_ascii=False, indent=2)


def process_transfers(from_year: str, to_year: str):
    base_path = f"unified_transfers/{from_year}_{to_year}"
    os.makedirs("docs/csv", exist_ok=True)

    countries_file = "docs/csv/countries.csv"
    if not os.path.exists(countries_file):
        print("❌ Arquivo docs/csv/countries.csv não encontrado.")
        sys.exit(1)

    country_map = load_countries_map(countries_file)

    output_players = []
    output_transfers = []
    clubs_dict = {}

    transfer_id = 1

    for file_name in os.listdir(base_path):
        if not file_name.endswith(".json"):
            continue

        players = []
        file_path = os.path.join(base_path, file_name)
        with open(file_path, "r", encoding="utf-8") as f:
            players = json.load(f)

            for player in players:
                player_id = player["player_id"]
                transfers = player.get("transfers", [])

                country_id = country_map.get(player["nationality"])

                if not country_id:
                    print(f"⚠️ País '{country_name}' não encontrado em countries.csv")
                    exit(1)

                output_players.append({
                    "id": player_id,
                    "country_id": country_id
                })

                for t in transfers:
                    transfer_entry = {
                        "id": transfer_id,
                        "player_id": player_id,
                        "club_from": t["from"]["club_id"],
                        "club_to": t["to"]["club_id"],
                        "fee_eur": t["fee_eur"],
                        "is_loan": t["is_loan"],
                        "season": t["season"]
                    }

                    output_transfers.append(transfer_entry)

                    for side in ["from", "to"]:
                        club = t[side]
                        country_name = club["country"]
                        country_id = country_map.get(country_name)

                        if not country_id:
                            print(f"⚠️ País '{country_name}' não encontrado em countries.csv")
                            exit(1)

                        clubs_dict[club["club_id"]] = {
                            "id": club["club_id"],
                            "name": club["club_name"],
                            "country_id": country_id
                        }

                    transfer_id += 1

    save_csv("docs/csv/players.csv", ["id", "country_id"], output_players)
    save_csv("docs/csv/transfers.csv", ["id", "player_id", "club_from", "club_to", "fee_eur", "is_loan", "season"], output_transfers)
    save_csv("docs/csv/clubs.csv", ["id", "name", "country_id"], clubs_dict.values())

    print(f"✅ Processado! {len(output_players)} jogadores, {len(output_transfers)} transferências, {len(clubs_dict)} clubes.")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print("Uso: python format_unified_data.py <FROM_YEAR> <TO_YEAR>")
        sys.exit(1)

    from_year = sys.argv[1]
    to_year = sys.argv[2]
    process_transfers(from_year, to_year)
