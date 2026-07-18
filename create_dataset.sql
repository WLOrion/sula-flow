LOAD CSV FROM "/import/countries.csv" WITH HEADER AS row
CREATE (:Country {
  id: toInteger(row.id),
  name: row.name,
  continent: row.continent
});

LOAD CSV FROM "/import/clubs.csv" WITH HEADER AS row
CREATE (:Club {
  id: toInteger(row.id),
  name: row.name,
  country_id: toInteger(row.country_id)
});

LOAD CSV FROM "/import/players.csv" WITH HEADER AS row
CREATE (:Player {
  id: toInteger(row.id),
  country_id: toInteger(row.country_id)
});

LOAD CSV FROM "/import/transfers.csv" WITH HEADER AS row
CREATE (:Transfer {
  id: toInteger(row.id),
  player_id: toInteger(row.player_id),
  club_from: toInteger(row.club_from),
  club_to: toInteger(row.club_to),
  fee_eur: toFloat(row.fee_eur),
  is_loan: row.is_loan = "True",
  season: row.season
});

MATCH (t:Transfer), (p:Player)
WHERE t.player_id = p.id
MERGE (t)-[:HAS_PLAYER]->(p);

MATCH (t:Transfer), (c:Club)
WHERE t.club_from = c.id
MERGE (t)-[:FROM_CLUB]->(c);

MATCH (t:Transfer), (c:Club)
WHERE t.club_to = c.id
MERGE (t)-[:TO_CLUB]->(c);

MATCH (p:Player), (c:Country)
WHERE p.country_id = c.id
MERGE (p)-[:NATIONALITY]->(c);

MATCH (cl:Club), (c:Country)
WHERE cl.country_id = c.id
MERGE (cl)-[:LOCATED_IN]->(c);

MATCH (t:Transfer)-[:FROM_CLUB]->(c_from:Club),
  (t)-[:TO_CLUB]->(c_to:Club)
WHERE NOT c_from.id IN [515, 123, 75, 2077, 2113, 12604]
  AND NOT c_to.id IN [515, 123, 75, 2077, 2113, 12604]
MERGE (c_from)-[:CONNECTED_TO]->(c_to);