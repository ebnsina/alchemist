import { searchRegions, REGIONS } from './regions';

// One runnable check: the search is the only logic here, and a silent mismatch means
// somebody cannot find their own region.
function run() {
	const eq = (got: unknown, want: unknown, what: string) => {
		const a = JSON.stringify(got);
		const b = JSON.stringify(want);
		if (a !== b) throw new Error(`${what}: got ${a}, want ${b}`);
	};

	eq(searchRegions('').length, REGIONS.length, 'empty query returns everything');
	eq(searchRegions('mumbai').map((r) => r.id), ['ap-south-1'], 'finds by city');
	eq(searchRegions('ap-south-1').map((r) => r.id), ['ap-south-1'], 'finds by exact code');
	eq(searchRegions('MUMBAI').map((r) => r.id), ['ap-south-1'], 'case insensitive');
	eq(searchRegions('sao paulo').map((r) => r.id), ['sa-east-1'], 'folds accents');
	eq(searchRegions('europe').length, 5, 'finds by area');
	eq(searchRegions('  london  ').map((r) => r.id), ['eu-west-2'], 'trims the query');
	eq(searchRegions('atlantis'), [], 'no match is empty, not everything');

	console.log('regions: all checks pass');
}

run();
