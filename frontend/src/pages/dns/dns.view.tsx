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

const results: DnsResponse[] = [];

const Dns: FC = () => {
    const [result, setResult] = useState<DnsResponse | null>(null);
    const [errors, setErrors] = useState<Record<DnsType, Error>>({} as Record<DnsType, Error>);

    const empty =
        result && !Object.values(result).find((record) => !!record.length);

    async function fetch(host: string, type: DnsType, trace: boolean) {
        await request<DnsResponse>(
            endpoint('/dns/:type/:host', { host, type, trace })
        ).then(res => {
            results.push(res);
            setResult((prev) => {
                if (prev) {
                    return {
                        ...prev,
                        [type]: [
                            ...(prev[type] || []),
                            ...(res[type] || [])
                        ]
                    };
                }

                return res;
            });
        }).catch((err) => {
            if (err instanceof Error) {
                setErrors((prev) => {
                    const newErrors = { ...prev };
                    newErrors[type] = err;
                    return newErrors;
                });
            } else {
                setErrors((prev) => {
                    const newErrors = { ...prev };
                    newErrors[type] = new Error('Unknown error');
                    return newErrors;
                });
            }
        });
    }

    return (
        <Stack spacing={6}>
            <Heading size="lg">DNS</Heading>

            <Card>
                <DnsForm
                    disabled={false}
                    onSubmit={async ({ host, type, trace }) => {
                        setResult(null);
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

                            // Fetch all types
                            for (const t of types) {
                                await fetch(host, t, trace);
                            }
                            return;
                        }
                        await fetch(host, type, trace);
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
                                    <DnsTable record={record} />
                                </Card>
                            </Fragment>
                        );
                    })}
                {(Object.entries(errors).map(([type, error]) => (
                        <Fragment key={type}>
                            <Box>
                                <Tag colorScheme="red" size="lg">
                                    {type}
                                </Tag>
                            </Box>
                            <Card overflowX="auto">
                                <Text color="red.500">
                                    {error.message}
                                </Text>
                            </Card>
                        </Fragment>
                    )))
                }
            </Stack>
        </Stack>
    );
};

export default Dns;