/**
 * A small raw-WebGL stand-in for the ShaderGradient "plane" preset.
 *
 * The published component is React and pulls three.js plus react-three-fiber —
 * around 150 KB gzipped, on a page whose audience is a cheap Android over 4G. This
 * reproduces the look directly: a subdivided plane displaced by 3D simplex noise,
 * three colours mixed across the surface, lambert shading from a single light, and
 * a grain pass. It is the same uniforms, without the framework.
 *
 * Returns null when WebGL is unavailable, so the caller keeps its CSS fallback.
 *
 * @param {HTMLCanvasElement} canvas
 * @param {{color1:string,color2:string,color3:string,uAmplitude:number,uDensity:number,
 *   uFrequency:number,uSpeed:number,uStrength:number,cDistance:number,fov:number,
 *   positionX:number,positionY:number,positionZ:number,rotationX:number,rotationY:number,
 *   rotationZ:number,brightness:number,reflection:number,grain:boolean}} cfg
 */
export function createShaderGradient(canvas, cfg) {
	const gl = canvas.getContext('webgl', {
		alpha: true,
		antialias: false,
		depth: true,
		powerPreference: 'low-power'
	});
	if (!gl) return null;

	const VERT = `
precision highp float;
attribute vec2 aXY;
uniform mat4 uProj, uView;
uniform float uTime, uAmplitude, uDensity, uFrequency, uStrength;
varying vec2 vUv;
varying float vDisp;
varying vec3 vPos;

vec3 mod289(vec3 x){return x-floor(x*(1./289.))*289.;}
vec4 mod289(vec4 x){return x-floor(x*(1./289.))*289.;}
vec4 permute(vec4 x){return mod289(((x*34.)+1.)*x);}
vec4 taylorInvSqrt(vec4 r){return 1.79284291400159-0.85373472095314*r;}
float snoise(vec3 v){
  const vec2 C=vec2(1./6.,1./3.); const vec4 D=vec4(0.,.5,1.,2.);
  vec3 i=floor(v+dot(v,C.yyy)); vec3 x0=v-i+dot(i,C.xxx);
  vec3 g=step(x0.yzx,x0.xyz); vec3 l=1.-g; vec3 i1=min(g.xyz,l.zxy); vec3 i2=max(g.xyz,l.zxy);
  vec3 x1=x0-i1+C.xxx; vec3 x2=x0-i2+C.yyy; vec3 x3=x0-D.yyy;
  i=mod289(i);
  vec4 p=permute(permute(permute(i.z+vec4(0.,i1.z,i2.z,1.))+i.y+vec4(0.,i1.y,i2.y,1.))+i.x+vec4(0.,i1.x,i2.x,1.));
  float n_=1./7.; vec3 ns=n_*D.wyz-D.xzx;
  vec4 j=p-49.*floor(p*ns.z*ns.z);
  vec4 x_=floor(j*ns.z); vec4 y_=floor(j-7.*x_);
  vec4 x=x_*ns.x+ns.yyyy; vec4 y=y_*ns.x+ns.yyyy; vec4 h=1.-abs(x)-abs(y);
  vec4 b0=vec4(x.xy,y.xy); vec4 b1=vec4(x.zw,y.zw);
  vec4 s0=floor(b0)*2.+1.; vec4 s1=floor(b1)*2.+1.; vec4 sh=-step(h,vec4(0.));
  vec4 a0=b0.xzyw+s0.xzyw*sh.xxyy; vec4 a1=b1.xzyw+s1.xzyw*sh.zzww;
  vec3 p0=vec3(a0.xy,h.x); vec3 p1=vec3(a0.zw,h.y); vec3 p2=vec3(a1.xy,h.z); vec3 p3=vec3(a1.zw,h.w);
  vec4 norm=taylorInvSqrt(vec4(dot(p0,p0),dot(p1,p1),dot(p2,p2),dot(p3,p3)));
  p0*=norm.x;p1*=norm.y;p2*=norm.z;p3*=norm.w;
  vec4 m=max(.6-vec4(dot(x0,x0),dot(x1,x1),dot(x2,x2),dot(x3,x3)),0.); m=m*m;
  return 42.*dot(m*m,vec4(dot(p0,x0),dot(p1,x1),dot(p2,x2),dot(p3,x3)));
}

float field(vec2 p){
  float t = uTime;
  float n = snoise(vec3(p * uDensity, t)) * uStrength;
  n += snoise(vec3(p * uDensity * uFrequency * 0.45, t * 1.3)) * uStrength * 0.22;
  return n * uAmplitude * 0.16;
}

void main(){
  vUv = aXY * 0.5 + 0.5;
  float d = field(aXY);
  vDisp = d;
  vec3 pos = vec3(aXY, d);
  vPos = pos;
  gl_Position = uProj * uView * vec4(pos, 1.0);
}`;

	const FRAG = `
precision highp float;
uniform vec3 uC1, uC2, uC3;
uniform float uBrightness, uReflection, uGrain, uTime;
varying vec2 vUv;
varying float vDisp;
varying vec3 vPos;

float hash(vec2 p){ return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

void main(){
  // Three colours laid across the surface, pushed around by the displacement, the
  // way the plane preset blends them.
  float a = clamp(vUv.x * 0.85 + vDisp * 1.4 + 0.08, 0.0, 1.0);
  float b = clamp(vUv.y * 0.9 - vDisp * 1.1, 0.0, 1.0);
  vec3 col = mix(uC1, uC2, smoothstep(0.0, 1.0, a));
  col = mix(col, uC3, smoothstep(0.15, 1.0, b) * 0.85);

  // lightType: 3d — one directional light, plus a weak specular for reflection.
  vec3 n = normalize(vec3(-dFdx(vDisp) * 40.0, -dFdy(vDisp) * 40.0, 1.0));
  vec3 l = normalize(vec3(-0.35, 0.55, 0.75));
  float lam = 0.62 + 0.38 * max(dot(n, l), 0.0);
  float spec = pow(max(dot(reflect(-l, n), vec3(0.0, 0.0, 1.0)), 0.0), 18.0) * uReflection;
  col = col * lam * uBrightness + spec;

  col += (hash(gl_FragCoord.xy + uTime) - 0.5) * uGrain;
  gl_FragColor = vec4(clamp(col, 0.0, 1.0), 1.0);
}`;

	const ext = gl.getExtension('OES_standard_derivatives');
	const compile = (type, src) => {
		const sh = gl.createShader(type);
		gl.shaderSource(sh, type === gl.FRAGMENT_SHADER && ext ? '#extension GL_OES_standard_derivatives : enable\n' + src : src);
		gl.compileShader(sh);
		return gl.getShaderParameter(sh, gl.COMPILE_STATUS) ? sh : null;
	};
	const vs = compile(gl.VERTEX_SHADER, VERT);
	const fs = compile(gl.FRAGMENT_SHADER, FRAG);
	if (!vs || !fs) return null;
	const prog = gl.createProgram();
	gl.attachShader(prog, vs);
	gl.attachShader(prog, fs);
	gl.linkProgram(prog);
	if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) return null;
	gl.useProgram(prog);

	// Subdivided plane, indexed.
	const SEG = 128;
	const SIZE = 7.0;
	const verts = new Float32Array((SEG + 1) * (SEG + 1) * 2);
	let k = 0;
	for (let y = 0; y <= SEG; y++)
		for (let x = 0; x <= SEG; x++) {
			verts[k++] = (x / SEG - 0.5) * SIZE;
			verts[k++] = (y / SEG - 0.5) * SIZE;
		}
	const idx = new Uint32Array(SEG * SEG * 6);
	const u32 = gl.getExtension('OES_element_index_uint');
	let m = 0;
	for (let y = 0; y < SEG; y++)
		for (let x = 0; x < SEG; x++) {
			const i = y * (SEG + 1) + x;
			idx[m++] = i; idx[m++] = i + 1; idx[m++] = i + SEG + 1;
			idx[m++] = i + 1; idx[m++] = i + SEG + 2; idx[m++] = i + SEG + 1;
		}
	if (!u32) return null;

	const vbo = gl.createBuffer();
	gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
	gl.bufferData(gl.ARRAY_BUFFER, verts, gl.STATIC_DRAW);
	const loc = gl.getAttribLocation(prog, 'aXY');
	gl.enableVertexAttribArray(loc);
	gl.vertexAttribPointer(loc, 2, gl.FLOAT, false, 0, 0);
	const ibo = gl.createBuffer();
	gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo);
	gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, idx, gl.STATIC_DRAW);

	const U = (n) => gl.getUniformLocation(prog, n);
	const uProj = U('uProj'), uView = U('uView'), uTime = U('uTime');
	const hex = (h) => [1, 3, 5].map((i) => parseInt(h.slice(i, i + 2), 16) / 255);
	gl.uniform3fv(U('uC1'), hex(cfg.color1));
	gl.uniform3fv(U('uC2'), hex(cfg.color2));
	gl.uniform3fv(U('uC3'), hex(cfg.color3));
	gl.uniform1f(U('uAmplitude'), cfg.uAmplitude);
	gl.uniform1f(U('uDensity'), cfg.uDensity);
	gl.uniform1f(U('uFrequency'), cfg.uFrequency);
	gl.uniform1f(U('uStrength'), cfg.uStrength);
	gl.uniform1f(U('uBrightness'), cfg.brightness);
	gl.uniform1f(U('uReflection'), cfg.reflection);
	gl.uniform1f(U('uGrain'), cfg.grain ? 0.055 : 0);

	const rad = (d) => (d * Math.PI) / 180;
	function view() {
		// Rotations then position, then pull back by cDistance along the camera axis.
		const rx = rad(cfg.rotationX), ry = rad(cfg.rotationY), rz = rad(cfg.rotationZ);
		const cx = Math.cos(rx), sx = Math.sin(rx);
		const cy = Math.cos(ry), sy = Math.sin(ry);
		const cz = Math.cos(rz), sz = Math.sin(rz);
		// R = Rz * Ry * Rx, column-major
		const r = [
			cz * cy, sz * cy, -sy,
			cz * sy * sx - sz * cx, sz * sy * sx + cz * cx, cy * sx,
			cz * sy * cx + sz * sx, sz * sy * cx - cz * sx, cy * cx
		];
		return new Float32Array([
			r[0], r[1], r[2], 0,
			r[3], r[4], r[5], 0,
			r[6], r[7], r[8], 0,
			cfg.positionX, cfg.positionY, cfg.positionZ - cfg.cDistance, 1
		]);
	}
	function proj(aspect) {
		const f = 1 / Math.tan(rad(cfg.fov) / 2);
		const near = 0.1, far = 60;
		return new Float32Array([
			f / aspect, 0, 0, 0,
			0, f, 0, 0,
			0, 0, (far + near) / (near - far), -1,
			0, 0, (2 * far * near) / (near - far), 0
		]);
	}

	let w = 0, h = 0;
	function resize() {
		// Capped device pixel ratio: this is a background, and fill rate is battery.
		const dpr = Math.min(window.devicePixelRatio || 1, 1.5);
		const nw = Math.round(canvas.clientWidth * dpr);
		const nh = Math.round(canvas.clientHeight * dpr);
		if (nw === w && nh === h) return;
		w = canvas.width = nw;
		h = canvas.height = nh;
		gl.viewport(0, 0, w, h);
		gl.uniformMatrix4fv(uProj, false, proj(w / h || 1));
	}

	gl.enable(gl.DEPTH_TEST);
	gl.uniformMatrix4fv(uView, false, view());

	let raf = 0, t0 = 0, clock = 0, running = false;
	function frame(now) {
		if (!running) return;
		if (!t0) t0 = now;
		clock += ((now - t0) / 1000) * cfg.uSpeed;
		t0 = now;
		resize();
		gl.uniform1f(uTime, clock);
		gl.clearColor(0, 0, 0, 0);
		gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);
		gl.drawElements(gl.TRIANGLES, idx.length, gl.UNSIGNED_INT, 0);
		raf = requestAnimationFrame(frame);
	}

	return {
		start() {
			if (running) return;
			running = true;
			t0 = 0;
			raf = requestAnimationFrame(frame);
		},
		stop() {
			running = false;
			cancelAnimationFrame(raf);
		},
		destroy() {
			this.stop();
			const lose = gl.getExtension('WEBGL_lose_context');
			if (lose) lose.loseContext();
		}
	};
}
