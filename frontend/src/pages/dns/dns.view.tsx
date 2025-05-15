import { Box, Heading, Stack, Tag, Text } from '@chakra-ui/react';
import { FC, Fragment, useState } from 'react';
import endpoint from '../../api/endpoint';
import request from '../../api/request';
import Card from '../../common/card/card';
import DnsForm from './dns-form';
import DnsTable from './dns-table';
import { getDnsColor } from './dns.helpers';
import { DnsType } from './dns.schema';

type DnsRecord = {
    type: DnsType;
    ttl: number;
    priority?: number;
    name: string;
    dnsServer: string;
    value:
        | string
        | string[]
        | {
        expire: number;
        mbox: string;
        minttl: number;
        ns: string;
        refresh: number;
        retry: number;
        serial: number;
    };
};

export type DnsResponse = {
    [key in DnsType]: {
        server: string;
        records: DnsRecord[];
    }[];
};

const Dns: FC = () => {
    const [result, setResult] = useState<DnsResponse | null>(null);

    const empty =
        result && !Object.values(result).find((record) => !!record.length);

    return (
        <Stack spacing={6}>
            <Heading size="lg">DNS</Heading>

            <Card>
                <DnsForm
                    disabled={false}
                    onSubmit={async ({host, type, trace}) => {
                        if (type === DnsType.ANY) {
                            const types = [
                                DnsType.A,
                                DnsType.AAAA,
                                DnsType.CNAME,
                                DnsType.MX,
                                DnsType.NS,
                                DnsType.SOA,
                                DnsType.TXT,
                            ];

                            const results: DnsResponse[] = [];
                            const errors: string[] = [];

                            for (const t of types) {
                                try {
                                    const res = await request<DnsResponse>(
                                        endpoint('/dns/:type/:host', { host, type: t, trace })
                                    );
                                    results.push(res);
                                } catch (error) {
                                    errors.push(`Error fetching ${t} records: ${error}`);
                                }
                            }
                            const combinedResult: DnsResponse = results.reduce(
                                (acc, res) => {
                                    Object.keys(res).forEach((key) => {
                                        const dnsKey = key as DnsType;
                                        acc[dnsKey] = acc[dnsKey] || [];
                                        acc[dnsKey].push(...res[dnsKey]);
                                    });
                                    return acc;
                                },
                                {} as DnsResponse
                            );

                            if (errors.length > 0) {
                                for (const error of errors) {
                                    const errorType = error.split(' ')[2] as DnsType;
                                    combinedResult[errorType] = combinedResult[errorType] || [];
                                    combinedResult[errorType].push({
                                        records: [
                                            {
                                                type: errorType,
                                                ttl: 0,
                                                name: host,
                                                dnsServer: '',
                                                value: error,
                                            }
                                        ],
                                        server: '',
                                    });
                                }
                            }

                            const orderedResult = Object.keys(combinedResult)
                                .sort()
                                .reduce((acc, key) => {
                                    acc[key as DnsType] = combinedResult[key as DnsType];
                                    return acc;
                                }, {} as DnsResponse);

                            setResult(orderedResult);
                            return;
                        }
                        const result = await request<DnsResponse>(
                            endpoint('/dns/:type/:host', {host, type, trace})
                        );
                        setResult(result);
                    }}
                />
            </Card>

            <Stack spacing={6}>
                {result !== null && empty && (
                    <Card>
                        <Text color="gray.500">No DNS records were found.</Text>
                    </Card>
                )}

                {result !== null &&
                    !empty &&
                    (Object.keys(result) as DnsType[]).map((type) => {
                        const record = result[type];

                        if (!record.length) return null;

                        return (
                            <Fragment key={type}>
                                <Box>
                                    <Tag colorScheme={getDnsColor(type)} size="lg">
                                        {type}
                                    </Tag>
                                </Box>
                                <Card overflowX="auto">
                                    <DnsTable record={record}/>
                                </Card>
                            </Fragment>
                        );
                    })}
            </Stack>
        </Stack>
    )
        ;
};

export default Dns;
