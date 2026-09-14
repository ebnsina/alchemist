/*
 * Playback signature validation at the edge.
 *
 * This runs in the nginx worker so the control plane is never in the playback path:
 * a cache hit is served without any origin or database contact at all.
 *
 * The canonical string and truncation MUST match internal/platform/signing/signing.go
 * exactly. signing_parity_test.go asserts that.
 */
import crypto from 'crypto';

var SIGNATURE_LENGTH = 32;

// Keys arrive as an nginx variable: "kid:secret,kid:secret". Every listed key is
// accepted so rotation is not an outage; the control plane decides which one signs.
function loadKeyring(r) {
    var raw = r.variables.playback_keys || '';
    var ring = {};
    raw.split(',').forEach(function (entry) {
        var i = entry.indexOf(':');
        if (i > 0) {
            ring[entry.substring(0, i).trim()] = entry.substring(i + 1).trim();
        }
    });
    return ring;
}

function compute(secret, prefix, exp) {
    return crypto.createHmac('sha256', secret)
        .update(prefix + '|' + exp)
        .digest('hex')
        .substring(0, SIGNATURE_LENGTH);
}

// Constant-time compare: a timing oracle on the signature would let an attacker
// recover a valid one byte by byte.
function secureEqual(a, b) {
    if (a.length !== b.length) return false;
    var diff = 0;
    for (var i = 0; i < a.length; i++) {
        diff |= a.charCodeAt(i) ^ b.charCodeAt(i);
    }
    return diff === 0;
}

/*
 * The signature covers the asset prefix, not the individual file, so one token
 * authorizes the manifest plus every segment and key request beneath it.
 * /playback/{tenant}/{asset}/{file} -> /playback/{tenant}/{asset}
 */
function assetPrefix(uri) {
    var parts = uri.split('/');
    if (parts.length < 5 || parts[1] !== 'playback') return null;
    return '/' + parts[1] + '/' + parts[2] + '/' + parts[3];
}

function authorize(r) {
    var prefix = assetPrefix(r.uri);
    if (prefix === null) {
        r.return(403);
        return;
    }

    var exp = r.args.exp;
    var kid = r.args.kid;
    var sig = r.args.sig;
    if (!exp || !kid || !sig) {
        r.return(403);
        return;
    }

    var expNum = Number(exp);
    if (!Number.isInteger(expNum) || expNum <= 0) {
        r.return(403);
        return;
    }
    if (Math.floor(Date.now() / 1000) > expNum) {
        r.return(403);
        return;
    }

    var secret = loadKeyring(r)[kid];
    if (!secret) {
        r.return(403);
        return;
    }

    if (!secureEqual(sig, compute(secret, prefix, exp))) {
        r.return(403);
        return;
    }
    r.return(204);
}

export default { authorize, compute, assetPrefix, secureEqual };
