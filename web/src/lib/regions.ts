// The S3 regions a customer is likely to hold video in, with the city so the choice
// can be made on geography rather than on remembering that ap-south-1 is Mumbai.
// Latency to the viewer is what matters, and "Mumbai" is the part that says it.
export type Region = { id: string; city: string; area: string };

export const REGIONS: Region[] = [
	{ id: 'ap-south-1', city: 'Mumbai', area: 'Asia Pacific' },
	{ id: 'ap-southeast-1', city: 'Singapore', area: 'Asia Pacific' },
	{ id: 'ap-southeast-2', city: 'Sydney', area: 'Asia Pacific' },
	{ id: 'ap-southeast-3', city: 'Jakarta', area: 'Asia Pacific' },
	{ id: 'ap-northeast-1', city: 'Tokyo', area: 'Asia Pacific' },
	{ id: 'ap-northeast-2', city: 'Seoul', area: 'Asia Pacific' },
	{ id: 'ap-east-1', city: 'Hong Kong', area: 'Asia Pacific' },
	{ id: 'me-south-1', city: 'Bahrain', area: 'Middle East' },
	{ id: 'me-central-1', city: 'UAE', area: 'Middle East' },
	{ id: 'eu-west-1', city: 'Ireland', area: 'Europe' },
	{ id: 'eu-west-2', city: 'London', area: 'Europe' },
	{ id: 'eu-central-1', city: 'Frankfurt', area: 'Europe' },
	{ id: 'eu-north-1', city: 'Stockholm', area: 'Europe' },
	{ id: 'eu-south-1', city: 'Milan', area: 'Europe' },
	{ id: 'us-east-1', city: 'N. Virginia', area: 'North America' },
	{ id: 'us-east-2', city: 'Ohio', area: 'North America' },
	{ id: 'us-west-1', city: 'N. California', area: 'North America' },
	{ id: 'us-west-2', city: 'Oregon', area: 'North America' },
	{ id: 'ca-central-1', city: 'Canada', area: 'North America' },
	{ id: 'sa-east-1', city: 'São Paulo', area: 'South America' },
	{ id: 'af-south-1', city: 'Cape Town', area: 'Africa' }
];

// Matches the code, the city or the area, so "mumbai", "ap-south" and "asia" all find
// the same row. Accents are folded because "sao paulo" should find "São Paulo".
const fold = (s: string) =>
	s.normalize('NFD').replace(/\p{Diacritic}/gu, '').toLowerCase();

export function searchRegions(query: string): Region[] {
	const q = fold(query.trim());
	if (!q) return REGIONS;
	return REGIONS.filter(
		(r) => fold(r.id).includes(q) || fold(r.city).includes(q) || fold(r.area).includes(q)
	);
}
