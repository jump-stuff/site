<script lang="ts">
  import { onMount } from 'svelte';
  import { MeshLambertMaterial, FrontSide, DoubleSide } from 'three';
  import * as topojson from 'topojson-client';

  import bg from '$lib/assets/logo/bg.png';
  import mask from '$lib/assets/logo/mask.png';
  import overlay from '$lib/assets/logo/overlay.png';

  let globe: HTMLDivElement | undefined = $state();

  onMount(async () => {
    const Globe = (await import('globe.gl')).default;
    if (globe) {
      const globeJWC = new Globe(globe, {
        animateIn: true,
        waitForGlobeReady: true
      })
        .showGlobe(false)
        .showAtmosphere(false);

      fetch('//cdn.jsdelivr.net/npm/world-atlas/land-110m.json')
        .then((res) => res.json())
        .then((landTopo) => {
          globeJWC
            //@ts-ignore
            .polygonsData(topojson.feature(landTopo, landTopo.objects.land).features)
            .polygonCapMaterial(new MeshLambertMaterial({ color: '#70709D', side: DoubleSide }))
            .polygonSideColor(() => 'rgba(0,0,0,0)');
        });

      globeJWC.width(512);
      globeJWC.height(512);

      globeJWC.backgroundColor('#333343');

      globeJWC.controls().autoRotate = true;
      globeJWC.controls().autoRotateSpeed = 1;
      globeJWC.controls().enableZoom = false;
      globeJWC.controls().enablePan = false;
      globeJWC.controls().enableRotate = false;
    }
  });
</script>

<div
  class="group relative size-128 animate-[spin_120s_linear_infinite] select-none"
  style:clip-path="circle(50%)">
  <img class="absolute z-10 select-none" src={overlay} alt="" draggable="false" />
  <div
    class="mask-alpha mask-center mask-no-repeat transition-opacity select-none group-hover:opacity-0"
    style={`mask-image: url(${mask})`}
    draggable="false">
    <div class="scale-140 animate-[reverse-spin_120s_linear_infinite]">
      <div class="" bind:this={globe}></div>
    </div>
  </div>
</div>
