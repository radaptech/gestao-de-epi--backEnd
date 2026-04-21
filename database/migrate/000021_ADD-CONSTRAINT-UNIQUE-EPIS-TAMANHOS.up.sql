ALTER TABLE tamanhos_epis 
ADD CONSTRAINT uk_epi_tamanho_tenant 
UNIQUE (IdEpi, IdTamanho, tenant_id);