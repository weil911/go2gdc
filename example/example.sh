#!/bin/bash

#cd go2gdc/example


# from filter to answer

pathCmd=../bin/go2gdc
dirData=../example

fromStatus=filter
toStatus=answer

eg4count=12
for eg4id in $( seq ${eg4count} )
do
    eg=eg-4-${eg4id}--

    fromPath=${dirData}/${eg}.txt
    toPath=${dirData}/${eg}

    ${pathCmd} from=${fromStatus}:${fromPath} to=${toStatus}:${toPath}
done


# from answer to downloaded

cp ./example/eg-4-7--Project_TCGA-LUAD--Omicstype_gene_htseq_fpkm_uq--Sampletype_01.json ./example/eg-5-1--Project_TCGA-LUAD--Omicstype_gene_htseq_fpkm_uq--Sampletype_01.json
answer=./example/eg-5-1--Project_TCGA-LUAD--Omicstype_gene_htseq_fpkm_uq--Sampletype_01.json
downloaded=./example/eg-5-1--
../bin/go2gdc from=answer:${answer} to=downloaded:${downloaded}

cp ./example/eg-4-7--Project_TCGA-LUAD--Omicstype_gene_htseq_fpkm_uq--Sampletype_11.json ./example/eg-5-2--Project_TCGA-LUAD--Omicstype_gene_htseq_fpkm_uq--Sampletype_11.json
answer=./example/eg-5-2--Project_TCGA-LUAD--Omicstype_gene_htseq_fpkm_uq--Sampletype_11.json
downloaded=./example/eg-5-2--
../bin/go2gdc from=answer:${answer} to=downloaded:${downloaded}


