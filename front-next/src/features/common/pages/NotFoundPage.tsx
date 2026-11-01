import Link from 'next/link';
import { ArrowLeft, Home, Search } from 'lucide-react';
import { BrandLogo } from '@/components/common/BrandLogo';
import { Button } from '@/components/ui/Button';

export function NotFoundPage() {
  return (
    <div className="min-h-screen bg-brand-bg flex flex-col justify-between p-6 sm:p-10">
      <div className="flex justify-between items-center">
        <BrandLogo size="md" />
      </div>

      <div className="max-w-md mx-auto text-center space-y-6 py-12">
        <div className="h-24 w-24 rounded-3xl bg-brand-blue-50 text-brand-blue flex items-center justify-center mx-auto shadow-sm">
          <Search className="h-12 w-12" />
        </div>

        <div className="space-y-2">
          <h1 className="text-4xl sm:text-5xl font-black text-brand-navy tracking-tight">404</h1>
          <h2 className="text-xl font-bold text-brand-navy">Sahifa topilmadi</h2>
          <p className="text-sm text-brand-muted leading-relaxed">
            Siz qidirayotgan sahifa ko‘chirilgan, o‘chirilgan yoki mavjud emas.
          </p>
        </div>

        <div className="flex flex-col sm:flex-row items-center justify-center gap-3 pt-2">
          <Link href="/dashboard" className="w-full sm:w-auto">
            <Button variant="primary" size="md" fullWidth leftIcon={<Home className="h-4 w-4" />}>
              Bosh sahifa
            </Button>
          </Link>
          <Link href="/login" className="w-full sm:w-auto">
            <Button variant="outline" size="md" fullWidth leftIcon={<ArrowLeft className="h-4 w-4" />}>
              Login sahifasi
            </Button>
          </Link>
        </div>
      </div>

      <div className="text-center text-xs text-brand-muted">
        © {new Date().getFullYear()} QabulGo. Barcha huquqlar himoyalangan.
      </div>
    </div>
  );
}
